package pipe

import (
	"fmt"
	"testing"
	"time"
)

func TestSrcSine(t *testing.T) {
	c := NewController(10, 1, time.Millisecond)

	dataChan := SrcCosine(c, 100, 0)

	time.Sleep(time.Millisecond * 200)
	close(c.done)

	data := <-dataChan
	fmt.Println(data)

	data = <-dataChan
	fmt.Println(data)

	data = <-dataChan
	fmt.Println(data)
}
