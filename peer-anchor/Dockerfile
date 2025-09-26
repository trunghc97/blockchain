# Use the shared peer base image
FROM shared/peer-base:latest

# Set default environment variables
ENV PEER_NODE_TYPE=anchor
ENV PEER_NODE_ID=anchor-peer-1
ENV PEER_PORT=8084
ENV PEER_GRPC_PORT=9094

EXPOSE 8084 9094

CMD ["./peer"]