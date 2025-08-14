# Pinjol

Pinjol is a simple Go service built using the Echo framework. This project is designed to demonstrate a flat project structure and includes a Nix development environment for easy reproducibility.

## Features

- RESTful API with health and version endpoints
- Middleware for logging requests
- Dockerized for containerized deployment
- Nix-based development environment

## Prerequisites

- [Nix](https://nixos.org/): A tool for reproducible builds and development environments.

## Installing Nix

To install Nix, follow the instructions below or visit the [Zero to Nix](https://zero-to-nix.com/start/install/) website for more details.

Run the following command to install Nix:

```bash
curl --proto '=https' --tlsv1.2 -sSf -L https://install.determinate.systems/nix | sh -s -- install
```

After installation, open a new terminal session to ensure the `nix` command is available.

## Getting Started

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd pinjol
   ```

2. Enter the Nix development shell:
   ```bash
   nix develop
   ```

3. Run the application:
   ```bash
   make run
   ```

4. Run tests:
   ```bash
   make test
   ```

## Project Structure

```plaintext
.
├── main.go
├── handlers.go
├── middleware.go
├── config.go
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── flake.nix
├── Makefile
└── handlers_test.go
```

## License

This project is licensed under the MIT License.
