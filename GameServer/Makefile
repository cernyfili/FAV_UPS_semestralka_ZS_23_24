# Define Go command
GO = go

# Target executable
TARGET = GameServer

# Default: build and run the server
run:
	$(GO) run cmd/gameserver/main.go

# Build the executable
build:
	$(GO) build -ldflags="-s -w" -o $(TARGET) cmd/gameserver/main.go

# Clean up
clean:
	rm -f $(TARGET)
