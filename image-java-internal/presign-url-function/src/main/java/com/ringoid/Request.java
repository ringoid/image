package com.ringoid;

public class Request {
    private String bucket;
    private String key;

    public String getBucket() {
        return bucket;
    }

    public void setBucket(String bucket) {
        this.bucket = bucket;
    }

    public String getKey() {
        return key;
    }

    public void setKey(String key) {
        this.key = key;
    }

    @Override
    public String toString() {
        return "Request{" +
                "bucket='" + bucket + '\'' +
                ", key='" + key + '\'' +
                '}';
    }
}
