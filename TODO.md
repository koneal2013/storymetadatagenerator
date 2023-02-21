### Below is a list of features I would add to my solution:

- [x] Expose a Grpc and a Http Server listening on separate ports
- [x] Implement logging middleware for Http server
- [x] Expose Jaeger and Zipkin UI for tracing for both servers (Jaeger port: 16686. Zipkin port: 9411)
- [ ] Implement metrics with Prometheus UI using OpenTelemetry metrics collector
- [x] Create tests to validate the functionality of endpoints exposed by the Grpc and Http servers
- [ ] Create tests to validate the StoryMetadata API using go mocks (no internet connection required)
- [x] Implement mTLS in Grpc server
- [ ] Implement TLS in Http server (low priority as this service is likely to sit behind some sort of proxy or service mesh)
- [x] Provide Makefile with commands to generate code coverage analysis, run tests, and build/start docker containers
- [ ] Create an RPC call that returns a stream to make request with a larger result set more performant
  - A subset of the results will be pushed to the stream as soon as they are ready. This will allow the consuming service to process the results as they become available
- [ ] Create test to validate the Agent (orchestrator for spin up, graceful shutdown, and config management of the http and grpc servers)

