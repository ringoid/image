all: clean stage-deploy

build:
	@echo '--- Building getpresigned-image function ---'
	GOOS=linux go build lambda-get-presigned/get_presigned_url.go
	@echo '--- Building internal-handle-upload-image function ---'
	GOOS=linux go build lambda-handle-upload/internal_handle_upload.go
	@echo '--- Building get-own-photos-image function ---'
	GOOS=linux go build lambda-get-own-photos/get_own_photos.go
	@echo '--- Building delete-photo-image function ---'
	GOOS=linux go build lambda-delete-photo/delete_photo.go
	@echo '--- Building lambda-handle-task-image function ---'
	GOOS=linux go build lambda-handle-task/internal_handle_task.go lambda-handle-task/remove_photo.go lambda-handle-task/resize_photo.go

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

stage-deploy: stage-deploy-internal zip_lambda
	@echo '--- Build lambda stage ---'
	@echo 'Package template'
	sam package --template-file image-template.yaml --s3-bucket ringoid-cloudformation-template --output-template-file image-template-packaged.yaml
	@echo 'Deploy stage-image-stack'
	sam deploy --template-file image-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name stage-image-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=stage --no-fail-on-empty-changeset

clean:
	@echo '--- Delete old artifacts ---'
	rm -rf auth-template-packaged.yaml
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

