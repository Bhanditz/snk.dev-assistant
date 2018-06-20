package main

import (
	"fmt"
	"log"

	"github.com/ctava/tfcgo"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

func main() {

	s := op.NewScope()
	initXValue := op.Const(s.SubScope("x"), float32(2.0))

	_, initX, handleX, _ := tfcgo.VariableAndTensor(s, initXValue, float32(2.0))

	alpha := op.Const(s.SubScope("alpha"), float32(1.0))
	delta := op.Const(s.SubScope("delta"), float32(0.05))
	gradientDescent := op.ResourceApplyGradientDescent(s, handleX, alpha, delta)

	g, err := s.Finalize()
	if err != nil {
		log.Fatal(err)
	}

	sess, err := tf.NewSession(g, nil)
	defer sess.Close()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := sess.Run(nil, nil, []*tf.Operation{initX}); err != nil {
		log.Fatal(err)
	}

	err = g.WriteGraphAsText()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i <= 39; i++ {
		result, err := sess.Run(nil, []tf.Output{g.Operation("Variable/Read/ReadVariableOp").Output(0)}, []*tf.Operation{gradientDescent})
		if err != nil {
			//log.Fatal(s.Err())
			log.Fatal(err)
		}
		fmt.Printf("x: %12f \n", result[0].Value())
	}

}
