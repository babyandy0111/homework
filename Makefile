get-queue:
	aws sqs get-queue-url --queue-name user --region us-east-2

send-user:
	aws sqs send-message --message-body "{\"name\":\"jjjj\"}" --queue-url https://sqs.us-east-2.amazonaws.com/185271018684/
