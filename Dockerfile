# Pre-compiled lightweight image
FROM alpine:3.19
WORKDIR /app

# Copy the pre-compiled binary
COPY deploy-bins/qisur-api .

# Expose the default listening port
EXPOSE 8086

# Execute the microservice
CMD ["./qisur-api"]
