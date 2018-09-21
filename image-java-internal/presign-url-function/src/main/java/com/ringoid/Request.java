package com.ringoid;

public class Request {
    private Boolean warmUpRequest;
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

    public Boolean getWarmUpRequest() {
        return warmUpRequest;
    }

    public void setWarmUpRequest(Boolean warmUpRequest) {
        this.warmUpRequest = warmUpRequest;
    }

    @Override
    public String toString() {
        return "Request{" +
                "warmUpRequest='" + warmUpRequest + '\'' +
                ", bucket='" + bucket + '\'' +
                ", key='" + key + '\'' +
                '}';
    }
}
