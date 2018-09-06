package com.ringoid;

import com.amazonaws.HttpMethod;
import com.amazonaws.services.s3.AmazonS3;
import com.amazonaws.services.s3.AmazonS3ClientBuilder;
import com.amazonaws.services.s3.model.GeneratePresignedUrlRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.URI;
import java.net.URISyntaxException;
import java.net.URL;

public class PresignFunction {
    private final Logger log = LoggerFactory.getLogger(getClass());
    private static final String REGION = "eu-west-1";

    public Response handle(Request request) {
        log.info("handle request : {}", request);

        AmazonS3 s3Client = AmazonS3ClientBuilder.standard()
                .withRegion(REGION)
                .build();

        // Set the pre-signed URL to expire after 5 mins.
        java.util.Date expiration = new java.util.Date();
        long expTimeMillis = expiration.getTime();
        expTimeMillis += 1000 * 60 * 5;
        expiration.setTime(expTimeMillis);

        // Generate the pre-signed URL.
        System.out.println("start generate pre-signed url");
        GeneratePresignedUrlRequest generatePresignedUrlRequest =
                new GeneratePresignedUrlRequest(request.getBucket(), request.getKey())
                        .withMethod(HttpMethod.PUT)
                        .withExpiration(expiration);
        URL url = s3Client.generatePresignedUrl(generatePresignedUrlRequest);
        log.info("successfully generate pre-signed url : {}", url);
        URI uri = null;
        try {
            uri = url.toURI();
        } catch (URISyntaxException e) {
            log.error("error while taking uri from url : " + url, e);
            throw new RuntimeException(e);
        }
        log.info("successfully generate pre-signed uri : {}", uri.toString());

        Response resp = new Response();
        resp.setUri(uri.toString());
        log.info("return response to client : {}", resp);
        return resp;
    }

}
