.PHONY: proto
proto:
	buf generate --exclude-path proto/captcha \
				 --exclude-path proto/config \
				 --exclude-path proto/services