package blockdag

import (
	"container/list"
	"github.com/Qitmeer/qitmeer-lib/common/hash"
	"github.com/Qitmeer/qitmeer-lib/core/dag"
	"github.com/Qitmeer/qitmeer/core/merkle"
	"sort"
	"time"
)

// Some available DAG algorithm types
const (
	// A Scalable BlockDAG protocol
	phantom="phantom"

	// Phantom protocol V2
	phantom_v2="phantom_v2"

	// The order of all transactions is solely determined by the Tree Graph (TG)
	conflux="conflux"

	// Confirming Transactions via Recursive Elections
	spectre="spectre"
)

// Maximum number of the DAG tip
const MaxTips=100

// Maximum order of the DAG block
const MaxBlockOrder=uint(^uint32(0))

// MaxTipLayerGap
const MaxTipLayerGap=10

// StableConfirmations
const StableConfirmations=10

// It will create different BlockDAG instances
func NewBlockDAG(dagType string) IBlockDAG {
	switch dagType {
	case phantom:
		return &Phantom{}
	case phantom_v2:
		return &Phantom_v2{}
	case conflux:
		return &Conflux{}
	case spectre:
		return &Spectre{}
	}
	return nil
}

// The abstract inferface is used to build and manager DAG
type IBlockDAG interface {
	// Return the name
	GetName() string

	// This instance is initialized and will be executed first.
	Init(bd *BlockDAG) bool

	// Add a block
	AddBlock(ib IBlock) *list.List

	// Build self block
	CreateBlock(b *Block) IBlock

	// If the successor return nil, the underlying layer will use the default tips list.
	GetTipsList() []IBlock

	// Find block hash by order, this is very fast.
	GetBlockByOrder(order uint) *hash.Hash

	// Query whether a given block is on the main chain.
	IsOnMainChain(ib IBlock) bool

	// return the tip of main chain
	GetMainChainTip() IBlock

	// return the main parent in the parents
	GetMainParent(parents *dag.HashSet) IBlock
}

// The general foundation framework of DAG
type BlockDAG struct {
	// The genesis of block dag
	genesis hash.Hash

	// Use block hash to save all blocks with mapping
	blocks map[hash.Hash]IBlock

	// The total number blocks that this dag currently owned
	blockTotal uint

	// The terminal block is in block dag,this block have not any connecting at present.
	tips *dag.HashSet

	// This is time when the last block have added
	lastTime time.Time

	// The full sequence of dag, please note that the order starts at zero.
	order map[uint]*hash.Hash

	// Current dag instance used. Different algorithms work according to
	// different dag types config.
	instance IBlockDAG
}

// Acquire the name of DAG instance
func (bd *BlockDAG) GetName() string {
	return bd.instance.GetName()
}

// Initialize self, the function to be invoked at the beginning
func (bd *BlockDAG) Init(dagType string) IBlockDAG{
	bd.instance=NewBlockDAG(dagType)
	bd.instance.Init(bd)

	bd.lastTime=time.Unix(time.Now().Unix(), 0)

	return bd.instance
}

// This is an entry for update the block dag,you need pass in a block parameter,
// If add block have failure,it will return false.
func (bd *BlockDAG) AddBlock(b IBlockData) *list.List {
	if b == nil {
		return nil
	}
	if bd.HasBlock(b.GetHash()) {
		return nil
	}
	var parents []*hash.Hash
	if bd.GetBlockTotal() > 0 {
		parents = b.GetParents()
		if len(parents) == 0 {
			return nil
		}
		if !bd.HasBlocks(parents) {
			return nil
		}
	}
	if !bd.IsDAG(b) {
		return nil
	}
	//
	block := Block{id:bd.GetBlockTotal(),hash: *b.GetHash(), weight: 1, layer:0}
	if parents != nil {
		block.parents = dag.NewHashSet()
		var maxLayer uint=0
		for k, h := range parents {
			parent := bd.GetBlock(h)
			block.parents.AddPair(h,parent)
			parent.AddChild(&block)
			if k == 0 {
				block.mainParent = parent.GetHash()
			}

			if maxLayer==0 || maxLayer < parent.GetLayer() {
				maxLayer=parent.GetLayer()
			}
		}
		block.SetLayer(maxLayer+1)
	}

	if bd.blocks == nil {
		bd.blocks = map[hash.Hash]IBlock{}
	}
	ib:=bd.instance.CreateBlock(&block)
	bd.blocks[block.hash] = ib
	if bd.GetBlockTotal() == 0 {
		bd.genesis = *block.GetHash()
	}
	bd.blockTotal++
	//
	bd.updateTips(&block)
	//
	t:=time.Unix(b.GetTimestamp(), 0)
	if bd.lastTime.Before(t) {
		bd.lastTime=t
	}
	//
	return bd.instance.AddBlock(ib)
}

// Acquire the genesis block of chain
func (bd *BlockDAG) GetGenesis() IBlock {
	return bd.GetBlock(&bd.genesis)
}

// Acquire the genesis block hash of chain
func (bd *BlockDAG) GetGenesisHash() *hash.Hash {
	return &bd.genesis
}

// If the block is illegal dag,will return false.
func (bd *BlockDAG) IsDAG(b IBlockData) bool {
	return true
}

// Is there a block in DAG?
func (bd *BlockDAG) HasBlock(h *hash.Hash) bool {
	return bd.GetBlock(h) != nil
}

// Is there some block in DAG?
func (bd *BlockDAG) HasBlocks(hs []*hash.Hash) bool {
	for _, h := range hs {
		if !bd.HasBlock(h) {
			return false
		}
	}
	return true
}

// Acquire one block by hash
func (bd *BlockDAG) GetBlock(h *hash.Hash) IBlock {
	if h == nil {
		return nil
	}
	block, ok := bd.blocks[*h]
	if !ok {
		return nil
	}
	return block
}

// Total number of blocks
func (bd *BlockDAG) GetBlockTotal() uint {
	return bd.blockTotal
}

// return the terminal blocks, because there maybe more than one, so this is a set.
func (bd *BlockDAG) GetTips() *dag.HashSet {
	return bd.tips
}

// Acquire the tips array of DAG
func (bd *BlockDAG) GetTipsList() []IBlock {
	result:=bd.instance.GetTipsList()
	if result!=nil {
		return result
	}
	result=[]IBlock{}
	for k:=range bd.tips.GetMap(){
		result=append(result,bd.GetBlock(&k))
	}
	return result
}

// build merkle tree form current DAG tips
func (bd *BlockDAG) BuildMerkleTreeStoreFromTips() []*hash.Hash {
	parents:=bd.GetTips().SortList(false)
	return merkle.BuildParentsMerkleTreeStore(parents)
}

// Refresh the dag tip whith new block,it will cause changes in tips set.
func (bd *BlockDAG) updateTips(b *Block) {
	if bd.tips == nil {
		bd.tips = dag.NewHashSet()
		bd.tips.AddPair(b.GetHash(),b)
		return
	}
	for k := range bd.tips.GetMap() {
		block := bd.GetBlock(&k)
		if block.HasChildren() {
			bd.tips.Remove(&k)
		}
	}
	bd.tips.AddPair(b.GetHash(),b)
}

// The last time is when add one block to DAG.
func (bd *BlockDAG) GetLastTime() *time.Time{
	return &bd.lastTime
}

// Return the full sequence array.
func (bd *BlockDAG) GetOrder() map[uint]*hash.Hash {
	return bd.order
}

// Obtain block hash by global order
func (bd *BlockDAG) GetBlockByOrder(order uint) *hash.Hash{
	result:=bd.instance.GetBlockByOrder(order)
	if result!=nil {
		return result
	}
	if order>=uint(len(bd.order)) {
		return nil
	}
	return bd.order[order]
}

// Return the last order block
func (bd *BlockDAG) GetLastBlock() IBlock{
	if bd.GetBlockTotal()==0 {
		return nil
	}
	result:=bd.GetBlockByOrder(bd.GetBlockTotal()-1)
	if result==nil {
		return nil
	}
	return bd.GetBlock(result)
}

// This function need a stable sequence,so call it before sorting the DAG.
// If the h is invalid,the function will become a little inefficient.
func (bd *BlockDAG) GetPrevious(h *hash.Hash) *hash.Hash{
	if h==nil {
		return nil
	}
	if h.IsEqual(bd.GetGenesisHash()) {
		return nil
	}
	b:=bd.GetBlock(h)
	if b==nil {
		return nil
	}
	if b.GetOrder()==0{
		return nil
	}
	return bd.GetBlockByOrder(b.GetOrder()-1)
}

// Returns a future collection of block. This function is a recursively called function
// So we should consider its efficiency.
func (bd *BlockDAG) GetFutureSet(fs *dag.HashSet, b IBlock) {
	children := b.GetChildren()
	if children == nil || children.IsEmpty() {
		return
	}
	for k:= range children.GetMap() {
		if !fs.Has(&k) {
			fs.Add(&k)
			bd.GetFutureSet(fs, bd.GetBlock(&k))
		}
	}
}

// Query whether a given block is on the main chain.
// Note that some DAG protocols may not support this feature.
func (bd *BlockDAG) IsOnMainChain(h *hash.Hash) bool {
	return bd.instance.IsOnMainChain(bd.GetBlock(h))
}

// return the tip of main chain
func (bd *BlockDAG) GetMainChainTip() IBlock {
	return bd.instance.GetMainChainTip()
}

// return the main parent in the parents
func (bd *BlockDAG) GetMainParent(parents *dag.HashSet) IBlock {
	return bd.instance.GetMainParent(parents)
}

// Return the layer of block,it is stable.
// You can imagine that this is the main chain.
func (bd *BlockDAG) GetLayer(h *hash.Hash) uint {
	return bd.GetBlock(h).GetLayer()
}

// Return current general description of the whole state of DAG
func (bd *BlockDAG) GetGraphState() *dag.GraphState {
	gs:=dag.NewGraphState()
	if bd.tips!=nil && !bd.tips.IsEmpty() {
		gs.GetTips().AddList(bd.tips.List())

		gs.SetLayer(0)
		for _,v:=range bd.tips.GetMap() {
			tip:=v.(*Block)
			if tip.GetLayer()>gs.GetLayer(){
				gs.SetLayer(tip.GetLayer())
			}
		}
	}
	gs.SetTotal(bd.GetBlockTotal())
	gs.SetMainHeight(bd.GetMainChainTip().GetHeight())
	return gs
}

// Locate all eligible block by current graph state.
func (bd *BlockDAG) LocateBlocks(gs *dag.GraphState,maxHashes uint) []*hash.Hash {
	if gs.IsExcellent(bd.GetGraphState()) {
		return nil
	}
	queue := []IBlock{}
	fs:=dag.NewHashSet()
	for _,v:=range bd.GetTips().GetMap(){
		ib:=v.(IBlock)
		queue=append(queue,ib)
		fs.AddPair(ib.GetHash(),ib)
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if gs.GetTips().Has(cur.GetHash()) {
			continue
		}
		needRec:=true
		if cur.HasChildren() {
			for _,v := range cur.GetChildren().GetMap() {
				ib:=v.(IBlock)
				if gs.GetTips().Has(ib.GetHash()) || !fs.Has(ib.GetHash()) {
					needRec=false
					break
				}
			}
		}
		if needRec {
			fs.AddPair(cur.GetHash(),cur)
			if cur.HasParents() {
				for _,v := range cur.GetParents().GetMap() {
					ib:=v.(IBlock)
					if fs.Has(ib.GetHash()) {
						continue
					}
					queue=append(queue,ib)
				}
			}
		}
	}

	fsSlice:=BlockSlice{}
	for _,v:=range fs.GetMap(){
		ib:=v.(IBlock)
		fsSlice=append(fsSlice,ib)
	}

	result:=[]*hash.Hash{}
	if len(fsSlice)>=2 {
		sort.Sort(fsSlice)
	}

	for i:=0;i<len(fsSlice) ;i++  {
		if i>=int(maxHashes) {
			break
		}
		result=append(result,fsSlice[i].GetHash())
	}
	return result
}

// Judging whether block is the virtual tip that it have not future set.
func isVirtualTip(b IBlock, futureSet *dag.HashSet, anticone *dag.HashSet, children *dag.HashSet) bool {
	for k:= range children.GetMap() {
		if k.IsEqual(b.GetHash()) {
			return false
		}
		if !futureSet.Has(&k) && !anticone.Has(&k) {
			return false
		}
	}
	return true
}

// This function is used to GetAnticone recursion
func (bd *BlockDAG) recAnticone(b IBlock, futureSet *dag.HashSet, anticone *dag.HashSet, h *hash.Hash) {
	if h.IsEqual(b.GetHash()) {
		return
	}
	node:=bd.GetBlock(h)
	children := node.GetChildren()
	needRecursion := false
	if children == nil || children.Size() == 0 {
		needRecursion = true
	} else {
		needRecursion = isVirtualTip(b, futureSet, anticone, children)
	}
	if needRecursion {
		if !futureSet.Has(h) {
			anticone.Add(h)
		}
		parents := node.GetParents()

		//Because parents can not be empty, so there is no need to judge.
		for k:= range parents.GetMap() {
			bd.recAnticone(b, futureSet, anticone, &k)
		}
	}
}

// This function can get anticone set for an block that you offered in the block dag,If
// the exclude set is not empty,the final result will exclude set that you passed in.
func (bd *BlockDAG) GetAnticone(b IBlock, exclude *dag.HashSet) *dag.HashSet {
	futureSet := dag.NewHashSet()
	bd.GetFutureSet(futureSet, b)
	anticone := dag.NewHashSet()
	for k:= range bd.tips.GetMap() {
		bd.recAnticone(b, futureSet, anticone, &k)
	}
	if exclude != nil {
		anticone.Exclude(exclude)
	}
	return anticone
}

// Sort block by id
func (bd *BlockDAG) SortBlock(src []*hash.Hash) []*hash.Hash {
	if len(src)<=1 {
		return src
	}
	srcBlockS:=BlockSlice{}
	for i:=0;i<len(src) ;i++  {
		ib:=bd.GetBlock(src[i])
		if ib!=nil {
			srcBlockS=append(srcBlockS,ib)
		}
	}
	if len(srcBlockS)>=2 {
		sort.Sort(srcBlockS)
	}
	result:=[]*hash.Hash{}
	for i:=0;i<len(srcBlockS) ;i++  {
		result=append(result,srcBlockS[i].GetHash())
	}
	return result
}

// GetConfirmations
func (bd *BlockDAG) GetConfirmations(h *hash.Hash) uint {
	block:=bd.GetBlock(h)
	if block == nil {
		return 0
	}
	mainTip:=bd.GetMainChainTip()
	if bd.IsOnMainChain(h) {
		return mainTip.GetHeight()-block.GetHeight()
	}
	if !block.HasChildren() {
		return 0
	}
	//
	queue := []IBlock{}
	queue=append(queue,block)

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if bd.IsOnMainChain(cur.GetHash()) {
			return 1+mainTip.GetHeight()-cur.GetHeight()
		}
		if !cur.HasChildren() {
			return 0
		}else {
			for _,v := range cur.GetChildren().GetMap() {
				ib:=v.(IBlock)
				queue=append(queue,ib)
			}
		}
	}
	return 0
}