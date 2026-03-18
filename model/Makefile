VENV_DIR ?= .venv
PYTHON ?= python3
VENV_PYTHON := $(VENV_DIR)/bin/python
VENV_PIP := $(VENV_DIR)/bin/pip

.PHONY: train venv install clean-venv

$(VENV_PYTHON): requirements.txt
	$(PYTHON) -m venv $(VENV_DIR)
	$(VENV_PIP) install --upgrade pip
	$(VENV_PIP) install -r requirements.txt

venv: $(VENV_PYTHON)

install: venv

train: venv
	$(VENV_PYTHON) pipeline.py

clean-venv:
	rm -rf $(VENV_DIR)
