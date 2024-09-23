To get container running
1. docker import container-app.tar
2. docker import container-redis.tar
3. docker-compose up --build -d
App has started now


How to get the logs
1. After running the container run -> docker exec -it  <container name> /bin/sh
2. In the current directeory there will be a requests.log file which will have all the logs.
