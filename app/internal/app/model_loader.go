package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/adapters/tflite"
	"github.com/andrii-reshko/3c-informational-technical-support-lab/app/internal/domain"
)

type Model struct {
	Id        string
	Predictor *tflite.Predictor
	Scaler    *domain.ModelConfig
	Meta      *domain.ModelQuality
}

type Loader struct {
	models map[string]*Model
}

func NewLoader(modelsIds []string) *Loader {
	models, err := loadModelsById(modelsIds)
	if err != nil {
		panic(fmt.Sprintf("failed to load models: %v", err))
	}
	return &Loader{
		models: models,
	}
}

func (l *Loader) Get(id string) (*Model, error) {
	p, ok := l.models[id]
	if !ok {
		return nil, fmt.Errorf("model with id %s not found", id)
	}
	return p, nil
}

func loadJsonMeta(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func loadModelsById(ids []string) (map[string]*Model, error) {
	models := make(map[string]*Model)
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		modelPath := fmt.Sprintf("%s/models/%s_model.tflite", cwd, id)
		scalerPath := fmt.Sprintf("%s/models/%s_scaler.json", cwd, id)
		metaPath := fmt.Sprintf("%s/models/%s_meta.json", cwd, id)

		modelCfg := &domain.ModelConfig{}
		if err = loadJsonMeta(scalerPath, modelCfg); err != nil {
			return nil, err
		}

		meta := &domain.ModelQuality{}
		if err = loadJsonMeta(metaPath, meta); err != nil {
			return nil, err
		}

		predictor, err := tflite.NewPredictor(modelPath)
		if err != nil {
			return nil, err
		}

		models[id] = &Model{
			Id:        id,
			Predictor: predictor,
			Scaler:    modelCfg,
			Meta:      meta,
		}
	}
	return models, nil
}
