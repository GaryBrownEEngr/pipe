package pipe

import (
	"fmt"
	"testing"
)

func TestSrcSine(t *testing.T) {
	c := NewController(10, 1, 0.001)

	dataWire := SrcSine(c, 100, 0)
	wireEnd := dataWire.NewWireEnd()

	c.Start()
	data, _ := wireEnd.GetData()
	c.Stop()
	fmt.Println(data)

	data, _ = wireEnd.GetData()
	fmt.Println(data)

	data, _ = wireEnd.GetData()
	fmt.Println(data)

	data, _ = wireEnd.GetData()
	fmt.Println(data)
}
