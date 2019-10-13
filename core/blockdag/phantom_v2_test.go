package blockdag

import (
	"fmt"
	"testing"
)

func Test_V2_GetFutureSet(t *testing.T) {

	ibd, tbMap := InitBlockDAG(phantom_v2,"PH_fig2-blocks")
	if ibd==nil {
		t.FailNow()
	}

	//ph:=ibd.(*Phantom)
	anBlock := bd.GetBlock(tbMap[testData.PH_GetFutureSet.Input])
	bset := NewHashSet()
	bd.getFutureSet(bset,anBlock)
	fmt.Printf("Get %s future set：\n", testData.PH_GetFutureSet.Input)
	printBlockSetTag(bset,tbMap)
	//
	if !processResult(bset,changeToHashList(testData.PH_GetFutureSet.Output, tbMap)) {
		t.FailNow()
	}
}