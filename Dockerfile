FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-incident-io"]
COPY baton-incident-io /