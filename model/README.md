# Peak Forecasting Model Pipeline

This project trains a quantile regression MLP (TensorFlow) to predict short-term resource peaks from Azure VM trace data.

The pipeline currently trains one model for:
- Resource: `cpu_usage`
- Forecast horizon: `15` minutes

The training run exports:
- `*_model.tflite` (inference model)
- `*_scaler.json` (feature scaling parameters)
- `*_meta.json` (quality and model metadata)

## Project Overview

Main flow (`pipeline.py`):
1. Load and unify trace datasets from `data/` (or download from configured URLs).
2. Build lag, rolling, and time-based features.
3. Train an MLP with pinball loss for quantile outputs (P10/P90 interval).
4. Evaluate quality (`loss`, `mae`, `coverage`).
5. Export model artifacts to `artifacts/`.

Core modules:
- `src/data_loader.py`: dataset loading, mapping, normalization, concatenation.
- `src/features.py`: feature engineering and target peak generation.
- `src/model.py`: model architecture and quantile loss.
- `src/export.py`: TFLite + JSON artifact export.
- `src/config.py`: training constants and dataset configuration.

## Installation

### Prerequisites

- Python 3.10+ (recommended)
- `make`

### Option 1 (recommended): Makefile setup

From the `model/` directory:

```bash
make install
```

This will:
- create `.venv/`
- upgrade `pip`
- install dependencies from `requirements.txt`

### Option 2: Manual setup

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt
```

## Run the Pipeline

Recommended:

```bash
make train
```

Manual:

```bash
source .venv/bin/activate
python pipeline.py
```

## Output Artifacts

Artifacts are written to `artifacts/` with this naming pattern:
- `<node_id>_<resource>_<horizon>m_model.tflite`
- `<node_id>_<resource>_<horizon>m_scaler.json`
- `<node_id>_<resource>_<horizon>m_meta.json`

Example:
- `global_trace_cpu_usage_15m_model.tflite`
- `global_trace_cpu_usage_15m_scaler.json`
- `global_trace_cpu_usage_15m_meta.json`

## Cleanup

Remove the local virtual environment:

```bash
make clean-venv
```

