#!/bin/bash

# Monitoring script for Pinjol
# Usage: ./monitoring.sh {start|stop|restart|status|logs|test|clean}

COMMAND=$1
SERVICE=$2

case $COMMAND in
    start)
        echo "Starting monitoring services..."
        docker-compose -f docker/monitoring/docker-compose.yml up -d
        ;;
    stop)
        echo "Stopping monitoring services..."
        docker-compose -f docker/monitoring/docker-compose.yml down
        ;;
    restart)
        echo "Restarting monitoring services..."
        docker-compose -f docker/monitoring/docker-compose.yml restart
        ;;
    status)
        echo "Monitoring services status:"
        docker-compose -f docker/monitoring/docker-compose.yml ps
        ;;
    logs)
        if [ -n "$SERVICE" ]; then
            echo "Logs for $SERVICE:"
            docker-compose -f docker/monitoring/docker-compose.yml logs $SERVICE
        else
            echo "Logs for all services:"
            docker-compose -f docker/monitoring/docker-compose.yml logs
        fi
        ;;
    test)
        echo "Testing monitoring setup..."
        # Add test commands here
        ;;
    clean)
        echo "Cleaning monitoring services..."
        docker-compose -f docker/monitoring/docker-compose.yml down -v
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs|test|clean} [service]"
        exit 1
        ;;
esac
