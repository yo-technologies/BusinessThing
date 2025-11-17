FROM python:3.12-alpine AS final

WORKDIR /app

RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && update-ca-certificates

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY main.py .

RUN chown -R appuser:appuser /app

USER appuser

ENTRYPOINT ["python", "main.py"]
