# Image Service

### STAGE API ENDPOINT IS ``a9o3cw1o7j.execute-api.eu-west-1.amazonaws.com``
### PROD API ENDPOINT IS ````


### Get pre-signed url

* url ``https://{API ENDPOINT}/Prod/get_presigned``

POST request

Headers:

* Content-Type : application/json

Body:

    {
        "accessToken":"adasdasd-fadfs-sdffd",
        "extension":"jpg"
    }
    
    all parameters are required
    
 Response Body:
 
    {
        "errorCode":"",
        "errorMessage":"",
        "uri":"https://bla.com/bla"
    }
    
Possible errorCodes:

* InternalServerError
* InvalidAccessTokenClientError
* WrongRequestParamsClientError

## Analytics Events

1. IMAGE_USER_ASK_UPLOAD_PHOTO_LINK

* userId - string
* bucket - string
* photoKey - string
* unixTime - int
* eventType - string (IMAGE_USER_ASK_UPLOAD_PHOTO_LINK)

`{"userId":"aslkdl-asfmfa-asd","bucket":"origin-photo","photoKey":"aslkdl-asfmfa-asd","unixTime":1534338646,"eventType":"IMAGE_USER_ASK_UPLOAD_PHOTO_LINK"}`

2.

* userId - string
* bucket - string
* photoKey - string
* photoId - string
* photoType - string (origin in most cases)
* size - int
* unixTime - int
* eventType - string (IMAGE_USER_UPLOADED_PHOTO)

`{"userId":"aslkdl-asfmfa-asd","bucket":"origin-photo","photoKey":"aslkdl-asfmfa-asd","photoId":"aslkdl-asfmfa-asd","photoType":"origin","size":1200,"unixTime":1534338646,"eventType":"IMAGE_USER_ASK_UPLOAD_PHOTO_LINK"}`
