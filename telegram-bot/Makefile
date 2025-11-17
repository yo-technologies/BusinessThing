.PHONY: run build clean

run:
	python main.py

build:
	docker build -t telegram-bot .

clean:
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
