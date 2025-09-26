# Use the shared peer base image
FROM shared/peer-base:latest

# Copy any peer-specific configurations if needed
# COPY ./config ./config

# Set default environment variables
ENV PEER_NODE_TYPE=main-bank
ENV PEER_NODE_ID=main-bank-peer-1
ENV PEER_PORT=8082
ENV PEER_GRPC_PORT=9092

EXPOSE 8082 9092

CMD ["./peer"]