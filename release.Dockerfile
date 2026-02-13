FROM dhi.io/static:20250419
ARG APP_NAME
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/${APP_NAME} /beast-changelog-action
USER 0
ENTRYPOINT ["/beast-changelog-action"]
