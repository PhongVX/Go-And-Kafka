# Run Mongo
docker-compose up 
# Build Service
docker build --no-cache -t go_micro_mongo:latest .
# Run Service
docker run --name gmm -d --rm -p 8844:9090 --network microservice_network1 go_micro_mongo:latest