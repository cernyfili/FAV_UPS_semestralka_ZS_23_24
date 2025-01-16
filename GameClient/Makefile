# Define the virtual environment directory
VENV_DIR = venv

# Define the requirements file
REQUIREMENTS = requirements.txt

# Default target to create virtual environment, install dependencies, and run the application
all: install run

# Create virtual environment and ensure pip is installed
$(VENV_DIR)/scripts/activate: $(REQUIREMENTS)
	python3 -m venv $(VENV_DIR)
	$(VENV_DIR)/scripts/pip install --upgrade pip
	$(VENV_DIR)/scripts/pip install -r $(REQUIREMENTS)

# Install dependencies (activate the virtual environment first)
install: $(VENV_DIR)/scripts/activate
	@echo "Virtual environment and dependencies installed!"

# Run the application
run: $(VENV_DIR)/scripts/activate
	$(VENV_DIR)/scripts/python3 -m main

# Clean the virtual environment (optional)
clean:
	rm -rf $(VENV_DIR)

# Help target
help:
	@echo "Makefile for setting up Python environment"
	@echo "Usage:"
	@echo "  make          - Create virtual environment, install dependencies, and run the application"
	@echo "  make clean    - Remove virtual environment"