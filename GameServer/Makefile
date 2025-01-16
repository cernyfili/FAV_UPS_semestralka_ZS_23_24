# Define Go command
GO = go

# Target executable
TARGET = GameServer

# Default: tidy, build, and run the server
run: tidy build
	./$(TARGET)

# Tidy the module dependencies
tidy:
	$(GO) mod tidy -e

# Build the executable
build:
	$(GO) build -ldflags="-s -w" -o $(TARGET) cmd/gameserver/main.go

# Clean up
clean:
	rm -f $(TARGET)