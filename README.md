# Image Service

### STAGE API ENDPOINT IS ``wlzvkc7yf5.execute-api.eu-west-1.amazonaws.com``
### PROD API ENDPOINT IS ````


### Get pre-signed url

* url ``https://{API ENDPOINT}/Prod/get_presigned``

POST request

Headers:

* X-Ringoid-App-Version : 1       //int
* Content-Type : application/json

Body:

    {
        "accessToken":"adasdasd-fadfs-sdffd",
        "extension":"jpg",
        "clientPhotoId":"sldfjlskdfj--;lfk;lf"
    }
    
    all parameters are required
    
 Response Body:
 
    {
        "errorCode":"",
        "errorMessage":"",
        "uri":"https://bla.com/bla",
        "originPhotoId":"sdljfhsljkdhgsdkj",
        "clientPhotoId":"sldfjlskdfj--;lfk;lf"
    }
    
Possible errorCodes:

* InternalServerError
* InvalidAccessTokenClientError
* WrongRequestParamsClientError
* TooOldAppVersionClientError

### Get user's own photos

* url ``https://{API ENDPOINT}/Prod/get_own_photos?accessToken={ACCESS TOKEN}&resolution=480x640``

Headers:

* X-Ringoid-App-Version : 1       //int

GET request

Allowed Sizes:

* 480x640
* 720x960
* 1080x1440
* 1440x1920

Headers:

* X-Ringoid-App-Version : 1       //int
* Content-Type : application/json

 Response Body:
 
    {
        "errorCode":"",
        "errorMessage":"",
        "photos":[
            {"photoId":"12dd","originPhotoId":"sldkjflkjlkjlkjf","photoUri":"https://bla-bla.com/sss.jpg","likes":22},
            {"photoId":"13dd","originPhotoId":"mnbmvnbcxlsdfhwo","photoUri":"https://bla-bla.com/sss2.jpg","likes":0}
        ]
    }
    
Possible errorCodes:

* InternalServerError
* WrongRequestParamsClientError
* InvalidAccessTokenClientError
* TooOldAppVersionClientError

### Delete user's photo

* url ``https://{API ENDPOINT}/Prod/delete_photo``

POST request

Headers:

* X-Ringoid-App-Version : 1       //int
* Content-Type : application/json

Body:

    {
        "accessToken":"adasdasd-fadfs-sdffd",
        "photoId":"lsdkfjlskdjf-sdflksndfl"
    }
    
    all parameters are required
    
 Response Body:
 
    {
        "errorCode":"",
        "errorMessage":""
    }
    
Possible errorCodes:

* InternalServerError
* InvalidAccessTokenClientError
* WrongRequestParamsClientError
* TooOldAppVersionClientError


## Analytics Events

1. IMAGE_USER_ASK_UPLOAD_PHOTO_LINK

* userId - string
* bucket - string
* photoKey - string
* unixTime - int
* eventType - string (IMAGE_USER_ASK_UPLOAD_PHOTO_LINK)

`{"userId":"aslkdl-asfmfa-asd","bucket":"origin-photo","photoKey":"aslkdl-asfmfa-asd","unixTime":1534338646,"eventType":"IMAGE_USER_ASK_UPLOAD_PHOTO_LINK"}`

2. IMAGE_USER_UPLOAD_PHOTO

* userId - string
* bucket - string
* photoKey - string
* photoId - string
* photoType - string (origin in most cases)
* size - int
* unixTime - int
* eventType - string (IMAGE_USER_UPLOAD_PHOTO)

`{"userId":"aslkdl-asfmfa-asd","bucket":"origin-photo","photoKey":"aslkdl-asfmfa-asd","photoId":"aslkdl-asfmfa-asd","photoType":"origin","size":1200,"unixTime":1534338646,"eventType":"IMAGE_USER_UPLOAD_PHOTO"}`

3. IMAGE_USER_DELETE_PHOTO

* userId - string
* photoId - string
* unixTime - int
* eventType - string (IMAGE_USER_DELETE_PHOTO)

`{"userId":"aslkdl-asfmfa-asd","photoId":"aslkdl-asfmfa-asd","unixTime":1534338646,"eventType":"IMAGE_USER_DELETE_PHOTO"}`

4. IMAGE_REMOVE_TO_BIG_S3_OBJECT

* userId - string
* bucket - string
* key - string
* size - int
* unixTime - int
* eventType - string (IMAGE_REMOVE_TO_BIG_S3_OBJECT)

`{"userId":"aslkdl-asfmfa-asd","bucket":"bucket-name","key":"/foto.jpg","size":"1024","unixTime":1534338646,"eventType":"IMAGE_REMOVE_TO_BIG_S3_OBJECT"}`