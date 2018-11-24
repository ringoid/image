stage-all: clean stage-deploy
test-all: clean test-deploy
prod-all: clean prod-deploy

build:
	go get -u github.com/ringoid/commons
	@echo '--- Building getpresigned-image function ---'
	GOOS=linux go build lambda-get-presigned/get_presigned_url.go
	@echo '--- Building internal-handle-upload-image function ---'
	GOOS=linux go build lambda-handle-upload/internal_handle_upload.go
	@echo '--- Building get-own-photos-image function ---'
	GOOS=linux go build lambda-get-own-photos/get_own_photos.go
	@echo '--- Building delete-photo-image function ---'
	GOOS=linux go build lambda-delete-photo/delete_photo.go
	@echo '--- Building lambda-handle-task-image function ---'
	GOOS=linux go build lambda-handle-task/internal_handle_task.go lambda-handle-task/remove_photo.go lambda-handle-task/resize_photo.go lambda-handle-task/remove_s3_object.go
	@echo '--- Building warmup-image function ---'
	GOOS=linux go build lambda-warmup/warm_up.go
	@echo '--- Building lambda-handle-stream-image function ---'
	GOOS=linux go build lambda-handle-stream/handle_stream.go lambda-handle-stream/like_photo.go
	@echo '--- Building internal-get-images-image function ---'
	GOOS=linux go build lambda-internal-getimages/get_images.go
	@echo '--- Building internal-clean-db-image function ---'
	GOOS=linux go build lambda-clean-db/clean.go

test-deploy-internal:
	@echo '--- Build and deploy PresignFunction to TEST ---'
	cd image-java-internal && gradle build :presign-url-function:migratePresignFunctionToTest

stage-deploy-internal:
	@echo '--- Build and deploy PresignFunction to STAGE ---'
	cd image-java-internal && gradle build :presign-url-function:migratePresignFunctionToStage

prod-deploy-internal:
	@echo '--- Build and deploy PresignFunction to PROD ---'
	cd image-java-internal && gradle build :presign-url-function:migratePresignFunctionToStage

zip_lambda: build
	@echo '--- Zip getpresigned-image function ---'
	zip getpresigned-image.zip ./get_presigned_url
	@echo '--- Zip internal-handle-upload-image function ---'
	zip internal_handle_upload.zip ./internal_handle_upload
	@echo '--- Zip get-own-photos-image function ---'
	zip get_own_photos.zip ./get_own_photos
	@echo '--- Zip delete-photo-image function ---'
	zip delete_photo.zip ./delete_photo
	@echo '--- Zip internal-handle-task-image function ---'
	zip internal_handle_task.zip ./internal_handle_task
	@echo '--- Zip warmup-image function ---'
	zip warmup-image.zip ./warm_up
	@echo '--- Zip lambda-handle-stream-image function ---'
	zip handle_stream.zip ./handle_stream
	@echo '--- Zip internal-get-images-image function ---'
	zip get_images.zip ./get_images
	@echo '--- Zip internal-clean-db-image function ---'
	zip clean.zip ./clean

test-deploy: test-deploy-internal zip_lambda
	@echo '--- Build lambda test ---'
	@echo 'Package template 1 phase'
	sam package --template-file image-template_1.yaml --s3-bucket ringoid-cloudformation-template --output-template-file image-template-packaged.yaml
	@echo 'Deploy test-image-stack 1 phase'
	sam deploy --template-file image-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name test-image-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=test --no-fail-on-empty-changeset
	@echo 'Package template 2 phase'
	sam package --template-file image-template_2.yaml --s3-bucket ringoid-cloudformation-template --output-template-file image-template-packaged.yaml
	@echo 'Deploy test-image-stack 2 phase'
	sam deploy --template-file image-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name test-image-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=test --no-fail-on-empty-changeset
	@echo 'Package template 3 phase'
	sam package --template-file image-template_3.yaml --s3-bucket ringoid-cloudformation-template --output-template-file image-template-packaged.yaml
	@echo 'Deploy test-image-stack 3 phase'
	sam deploy --template-file image-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name test-image-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=test --no-fail-on-empty-changeset

stage-deploy: stage-deploy-internal zip_lambda
	@echo '--- Build lambda stage ---'
	@echo 'Package template 1 phase'
	sam package --template-file image-template_1.yaml --s3-bucket ringoid-cloudformation-template --output-template-file image-template-packaged.yaml
	@echo 'Deploy stage-image-stack 1 phase'
	sam deploy --template-file image-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name stage-image-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=stage --no-fail-on-empty-changeset
	@echo 'Package template 2 phase'
	sam package --template-file image-template_2.yaml --s3-bucket ringoid-cloudformation-template --output-template-file image-template-packaged.yaml
	@echo 'Deploy stage-image-stack 2 phase'
	sam deploy --template-file image-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name stage-image-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=stage --no-fail-on-empty-changeset
	@echo 'Package template 3 phase'
	sam package --template-file image-template_3.yaml --s3-bucket ringoid-cloudformation-template --output-template-file image-template-packaged.yaml
	@echo 'Deploy stage-image-stack 3 phase'
	sam deploy --template-file image-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name stage-image-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=stage --no-fail-on-empty-changeset

prod-deploy: prod-deploy-internal zip_lambda
	@echo '--- Build lambda prod ---'
	@echo 'Package template 1 phase'
	sam package --template-file image-template_1.yaml --s3-bucket ringoid-cloudformation-template --output-template-file image-template-packaged.yaml
	@echo 'Deploy prod-image-stack 1 phase'
	sam deploy --template-file image-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name prod-image-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=prod --no-fail-on-empty-changeset
	@echo 'Package template 2 phase'
	sam package --template-file image-template_2.yaml --s3-bucket ringoid-cloudformation-template --output-template-file image-template-packaged.yaml
	@echo 'Deploy prod-image-stack 2 phase'
	sam deploy --template-file image-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name prod-image-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=prod --no-fail-on-empty-changeset
	@echo 'Package template 3 phase'
	sam package --template-file image-template_3.yaml --s3-bucket ringoid-cloudformation-template --output-template-file image-template-packaged.yaml
	@echo 'Deploy prod-image-stack 3 phase'
	sam deploy --template-file image-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name prod-image-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=prod --no-fail-on-empty-changeset

clean:
	@echo '--- Delete old artifacts ---'
	rm -rf get_presigned_url
	rm -rf getpresigned-image.zip
	cd image-java-internal && gradle clean
	rm -rf internal_handle_upload.zip
	rm -rf internal_handle_upload
	rm -rf get_own_photos.zip
	rm -rf get_own_photos
	rm -rf delete_photo.zip
	rm -rf delete_photo
	rm -rf internal_handle_task.zip
	rm -rf internal_handle_task
	rm -rf warmup-image.zip
	rm -rf warm_up
	rm -rf handle_stream
	rm -rf handle_stream.zip
	rm -rf get_images.zip
	rm -rf get_images
	rm -rf clean.zip
	rm -rf clean

