#!/bin/bash

echo "Getting current directory..."
SCRIPT_DIR="$(pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
echo "Project root directory: $PROJECT_DIR"

echo "Cleaning old build files..."
rm -f $PROJECT_DIR/target/*.jar

echo "Building project with Maven..."
cd $PROJECT_DIR
./mvnw clean package -DskipTests

if [ ! -f $PROJECT_DIR/target/ests-0.0.1-SNAPSHOT.jar ]; then
  echo "Build failed, please check Maven configuration"
  exit 1
fi

echo "Stopping running application..."
PID=$(lsof -t -i:8080 || echo "")
if [ ! -z "$PID" ]; then
  kill -9 $PID
  echo "Killed process $PID"
  sleep 3
fi

echo "Starting application..."
cd $PROJECT_DIR
nohup java -jar target/ests-0.0.1-SNAPSHOT.jar > daemon.log 2>&1 &
echo 'Java application started, you can check logs in daemon.log'