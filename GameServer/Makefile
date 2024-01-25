# Define Go command and flags
GO = go
GOFLAGS = -ldflags="-s -w"

# Define the target executable
TARGET = GameServer.exe

# Default target: build the executable
all: $(TARGET)

# Rule to build the target executable
$(TARGET): cmd/gameserver/main.go
	$(GO) build $(GOFLAGS) -o $(TARGET) cmd/gameserver/main.go

# Clean target: remove the target executable
clean:
	rm -f $(TARGET)

# Run target: build and run the target executable
run: $(TARGET)
	./$(TARGET)

# Test target: run Go tests for the project
test:
	$(GO) test ./...