High-Level Overview
1. Technology Choice:
a. Go: I used Go because it's fast and can handle many requests at once without much overhead. Go's built-in tools make it easy to build an HTTP server that can manage thousands of requests per second, which was one of the key requirements.

b. Redis: I added Redis to help track and ensure that each request id is unique, even when the application is running multiple instances behind a load balancer. Redis is reliable and quick when it comes to storing and checking data across multiple instances

c. Docker: I containerized the application using Docker so it can be easily deployed and run in different environments. Docker also makes scaling the service easier by allowing multiple copies (instances) to run seamlessly.


2. Service Design:
a. API Endpoint (/api/verve/accept): The API accepts two parameters: an id (required) and an optional endpoint.

b . The service checks whether the id is unique, and if so, it counts the request. If an endpoint is provided, the service makes a separate HTTP request to that endpoint.

c. The goal is to ensure the service processes each unique id and handles concurrent requests without errors.

d. If endpoint is there, it'll make a call to that endpoint with param as unique_requests


3. Handling High Traffic:
a. Concurrency: Go’s goroutines are used to handle multiple requests at once. The service checks the uniqueness of each id using Redis, ensuring that only unique requests are counted, even when the app is distributed or handling high traffic.

b. Deduplication with Redis: Redis stores each id for one minute. If a request comes in with the same id within that time, it's ignored as a duplicate. This ensures that even when multiple instances are running, each id is counted only once.


4. Logging:
A log file is created where the number of unique requests received in each minute is logged.
This log file keeps track of how many unique ids were processed, making it easier to monitor the service's performance.

5. Firing POST Requests (Extension 1):
If an optional endpoint is provided, the service sends a POST request to that endpoint, including the count of unique requests in the current minute as a JSON payload.
The status of the POST request is logged to help with debugging and monitoring the success or failure of these requests.

6. Handling Load Balancers (Extension 2):
Redis for Deduplication: To make sure the id deduplication works even when multiple instances are behind a load balancer, Redis is used as a central place to store and check ids. This way, no matter which instance gets the request, the same id is never counted twice.

7. Future Extension: Distributed Streaming (Extension 3):
While the service logs unique requests to a file, it has commented code to send these logs to a distributed streaming service like Kafka.


Design Considerations
1. Scalability:
The use of Redis ensures that the service can scale horizontally, meaning multiple instances of the service can run in parallel without messing up the request count. Redis acts as the central store for request ids to ensure uniqueness.
2. Performance:
Go’s efficiency with concurrent requests ensures that the service can handle a high number of requests per second without slowing down.
Redis was chosen for its speed and atomic operations, which are perfect for checking if an id has been processed before.
3. Reliability:
Redis ensures that even in a distributed setup, the application correctly counts unique requests. If any instance crashes, Redis still maintains the id data, making it reliable.
The application writes logs regularly, ensuring that there’s a clear record of unique requests and any issues with the service
4. Flexibility:
The system is designed to allow easy modifications or future enhancements, such as using a distributed streaming service or changing how deduplication is handled. The application is modular and can adapt to new requirements.


How to get the logs
1. After running the container run -> docker exec -it  <container name> /bin/sh
2. In the current directeory there will be a requests.log file which will have all the log.





