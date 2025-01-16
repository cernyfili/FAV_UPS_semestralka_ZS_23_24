# Define the virtual environment directory
VENV_DIR = venv

# Define the requirements file
REQUIREMENTS = requirements.txt

# Default target to create virtual environment, install dependencies, and run the application
all: install run

# Create virtual environment and ensure pip is installed
$(VENV_DIR)/bin/activate: $(REQUIREMENTS)
	python -m venv --without-pip $(VENV_DIR)
	curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py
	$(VENV_DIR)/bin/python get-pip.py
	rm get-pip.py
	$(VENV_DIR)/bin/pip install --upgrade pip
	$(VENV_DIR)/bin/pip install -r $(REQUIREMENTS)

# Install dependencies (activate the virtual environment first)
install: $(VENV_DIR)/bin/activate
	@echo "Virtual environment and dependencies installed!"

# Run the application
run: $(VENV_DIR)/bin/activate
	$(VENV_DIR)/bin/python -m src.main

# Clean the virtual environment (optional)
clean:
	rm -rf $(VENV_DIR)

# Help target
help:
	@echo "Makefile for setting up Python environment"
	@echo "Usage:"
	@echo "  make          - Create virtual environment, install dependencies, and run the application"
	@echo "  make clean    - Remove virtual environment"