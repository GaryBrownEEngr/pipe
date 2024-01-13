package pipe

import (
	"fmt"
)

/*
The data-block should never be edited by a processing-unit that receives it.

A processing-unit should run until:
 - the controller done signal is received.
 - one of the input channels closes.

Then the input channels should be drained until they are closed. The output then will be closed.
*/

func GenericProcessingBlock_1In_1Out[T, U any](c *Controller, operation func([]T) []U, inWires ...*Wire[T]) *Wire[U] {
	// translate to a multiple output operation
	multiOutputOperation := func(in [][]T) [][]U {
		ret := operation(in[0])
		return [][]U{ret}
	}

	return GenericProcessingBlock_NIn_MOut(c, multiOutputOperation, 1, inWires...)[0]
}

func GenericProcessingBlock_NIn_1Out[T, U any](c *Controller, operation func([][]T) []U, inWires ...*Wire[T]) *Wire[U] {
	// translate to a multiple output operation
	multiOutputOperation := func(in [][]T) [][]U {
		ret := operation(in)
		return [][]U{ret}
	}

	return GenericProcessingBlock_NIn_MOut(c, multiOutputOperation, 1, inWires...)[0]
}

func GenericProcessingBlock_NIn_MOut[T, U any](c *Controller, operation func([][]T) [][]U, outputCount int, inWires ...*Wire[T]) []*Wire[U] {
	inWireEnds := make([]*WireEnd[T], len(inWires))
	for i := range inWires {
		inWireEnds[i] = inWires[i].NewWireEnd()
	}

	// outWires := NewWire[T](c.channelDepth)

	outWires := make([]*WireSource[U], outputCount)
	for i := range outWires {
		outWires[i] = NewWire[U](c.channelDepth)
	}

	go func() {
		inData := make([][]T, len(inWires))

		c.WaitForStart()

	MainLoop:
		for {
			select {
			case <-c.doneChan:
				break MainLoop
			default:
				// nothing
			}

			var ok bool
			// Get all the data for the next output block
			for i := range inWireEnds {
				inData[i], ok = inWireEnds[i].GetData()
				if !ok {
					break MainLoop
				}
				if len(inData[i]) != c.blockSize {
					panic(fmt.Sprint("bad block size", len(inData[i])))
				}
			}

			opResult := operation(inData)

			// Make sure the results are the correct size
			if len(opResult) != outputCount {
				panic(fmt.Sprint("result count from operation", len(opResult)))
			}
			for i := range opResult {
				if len(opResult[i]) != c.blockSize {
					panic(fmt.Sprint("bad block size from operation", len(opResult)))
				}
			}

			for i := range opResult {
				outWires[i].Publish(opResult[i])
			}
		}

		for i := range inWires {
			inWireEnds[i].Disconnect()
		}

		for i := range outWires {
			outWires[i].Stop()
		}
	}()

	ret := make([]*Wire[U], len(outWires))
	for i := range outWires {
		ret[i] = outWires[i].GetWire()
	}

	return ret
}

func TrimFirstN[T any](c *Controller, inChannel chan []T, n int) chan []T {
	outChan := make(chan []T, c.channelDepth)

	go func() {
		outDataLen := 0
		outData := make([]T, c.blockSize)

		c.WaitForStart()

	MainLoop:
		for {
			select {
			case <-c.doneChan:
				break MainLoop
			default:
				// nothing
			}

			inData, ok := <-inChannel
			if !ok {
				break MainLoop
			}
			if len(inData) != c.blockSize {
				panic(fmt.Sprint("bad block size", len(inData)))
			}

			for i := range inData {
				if n > 0 {
					n--
					continue
				}

				outData[outDataLen] = inData[i]
				outDataLen++
				if outDataLen == c.blockSize {
					outChan <- outData
					outDataLen = 0
					outData = make([]T, c.blockSize)
				}
			}
		}

		drainChannelTillClosed(inChannel)
		close(outChan)
	}()

	return outChan
}
