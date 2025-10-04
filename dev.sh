#!/bin/bash

# OneClick Development Scripts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
}

# Function to start development infrastructure
start_dev() {
    print_status "Starting OneClick development infrastructure..."
    
    check_docker
    
    # Copy environment file if it doesn't exist
    if [ ! -f .env ]; then
        print_status "Creating .env file from template..."
        cp env.dev.example .env
        print_warning "Please review and update .env file with your configuration"
    fi
    
    # Start infrastructure services
    print_status "Starting PostgreSQL and Redis..."
    docker-compose -f docker-compose.dev.yml up -d
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 10
    
    # Run migrations
    print_status "Running database migrations..."
    docker-compose -f docker-compose.dev.yml up migrate
    
    print_success "Development infrastructure is ready!"
    print_status "PostgreSQL: localhost:5433"
    print_status "Redis: localhost:6380"
    print_status "You can now run: go run cmd/server/main.go"
}

# Function to stop development infrastructure
stop_dev() {
    print_status "Stopping OneClick development infrastructure..."
    docker-compose -f docker-compose.dev.yml down
    print_success "Development infrastructure stopped!"
}

# Function to start full stack (infrastructure + backend)
start_full() {
    print_status "Starting full OneClick stack..."
    
    check_docker
    
    # Copy environment file if it doesn't exist
    if [ ! -f .env ]; then
        print_status "Creating .env file from template..."
        cp env.dev.example .env
        print_warning "Please review and update .env file with your configuration"
    fi
    
    # Start all services
    docker-compose up -d
    
    print_success "Full OneClick stack is ready!"
    print_status "Backend API: http://localhost:8080"
    print_status "PostgreSQL: localhost:5432"
    print_status "Redis: localhost:6379"
}

# Function to stop full stack
stop_full() {
    print_status "Stopping full OneClick stack..."
    docker-compose down
    print_success "Full OneClick stack stopped!"
}

# Function to start with monitoring
start_monitoring() {
    print_status "Starting OneClick with monitoring..."
    
    check_docker
    
    # Start all services including monitoring
    docker-compose --profile monitoring up -d
    
    print_success "OneClick with monitoring is ready!"
    print_status "Backend API: http://localhost:8080"
    print_status "Prometheus: http://localhost:9090"
    print_status "Grafana: http://localhost:3000 (admin/admin123)"
}

# Function to show logs
show_logs() {
    if [ "$1" = "backend" ]; then
        docker-compose logs -f backend
    elif [ "$1" = "postgres" ]; then
        docker-compose logs -f postgres
    elif [ "$1" = "redis" ]; then
        docker-compose logs -f redis
    else
        docker-compose logs -f
    fi
}

# Function to clean up
clean() {
    print_status "Cleaning up OneClick Docker resources..."
    
    # Stop and remove containers
    docker-compose down --remove-orphans
    docker-compose -f docker-compose.dev.yml down --remove-orphans
    
    # Remove volumes (WARNING: This will delete all data)
    read -p "Do you want to delete all data volumes? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker volume prune -f
        print_warning "All data volumes have been deleted!"
    fi
    
    print_success "Cleanup completed!"
}

# Function to show status
status() {
    print_status "OneClick Docker Services Status:"
    echo
    docker-compose ps
    echo
    docker-compose -f docker-compose.dev.yml ps
}

# Function to show help
show_help() {
    echo "OneClick Development Scripts"
    echo
    echo "Usage: $0 [COMMAND]"
    echo
    echo "Commands:"
    echo "  dev-start     Start development infrastructure (PostgreSQL + Redis)"
    echo "  dev-stop      Stop development infrastructure"
    echo "  full-start    Start full stack (infrastructure + backend)"
    echo "  full-stop     Stop full stack"
    echo "  monitoring    Start with monitoring (Prometheus + Grafana)"
    echo "  logs [service] Show logs (backend, postgres, redis, or all)"
    echo "  status        Show status of all services"
    echo "  clean         Clean up Docker resources"
    echo "  help          Show this help message"
    echo
    echo "Examples:"
    echo "  $0 dev-start"
    echo "  $0 logs backend"
    echo "  $0 monitoring"
}

# Main script logic
case "$1" in
    "dev-start")
        start_dev
        ;;
    "dev-stop")
        stop_dev
        ;;
    "full-start")
        start_full
        ;;
    "full-stop")
        stop_full
        ;;
    "monitoring")
        start_monitoring
        ;;
    "logs")
        show_logs "$2"
        ;;
    "status")
        status
        ;;
    "clean")
        clean
        ;;
    "help"|"--help"|"-h"|"")
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
