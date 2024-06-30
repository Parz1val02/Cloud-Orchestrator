#!/bin/sh

# Start the Celery worker with the app context
celery -A celery_linux.celery worker --loglevel=info
