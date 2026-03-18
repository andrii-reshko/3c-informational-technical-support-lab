package tflite

import (
	"fmt"

	"github.com/mattn/go-tflite"
)

type Predictor struct {
	model       *tflite.Model
	interpreter *tflite.Interpreter
}

func NewPredictor(modelPath string) (*Predictor, error) {
	model := tflite.NewModelFromFile(modelPath)
	if model == nil {
		return nil, fmt.Errorf("failed to load model")
	}

	options := tflite.NewInterpreterOptions()
	options.SetNumThread(2)

	interpreter := tflite.NewInterpreter(model, options)
	if interpreter == nil {
		return nil, fmt.Errorf("failed to create interpreter")
	}

	if status := interpreter.AllocateTensors(); status != tflite.OK {
		return nil, fmt.Errorf("failed to allocate tensors")
	}

	return &Predictor{model: model, interpreter: interpreter}, nil
}

func (p *Predictor) Predict(input []float32) ([]float32, error) {
	// 1. Заповнюємо вхідний тензор
	inputTensor := p.interpreter.GetInputTensor(0)
	inputTensor.SetFloat32s(input)

	// 2. Викликаємо інференс
	if status := p.interpreter.Invoke(); status != tflite.OK {
		return nil, fmt.Errorf("tflite invoke failed")
	}

	// 3. Отримуємо вихідний тензор
	// Зазвичай MLP для квантильної регресії має один вихідний тензор з декількома числами
	outputTensor := p.interpreter.GetOutputTensor(0)
	output := outputTensor.Float32s()

	if len(output) < 2 {
		return nil, fmt.Errorf("model returned less than 2 outputs, check your model architecture")
	}

	// Повертаємо весь слайс (наприклад, [0.45, 0.88])
	return output, nil
}
