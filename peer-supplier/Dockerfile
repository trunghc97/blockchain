# Use the shared peer base image
FROM shared/peer-base:latest

# Set default environment variables
ENV PEER_NODE_TYPE=supplier
ENV PEER_NODE_ID=supplier-peer-1
ENV PEER_PORT=8083
ENV PEER_GRPC_PORT=9093

EXPOSE 8083 9093

CMD ["./peer"]