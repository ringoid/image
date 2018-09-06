package com.ringoid;

public class Response {
    private String uri;

    public String getUri() {
        return uri;
    }

    public void setUri(String uri) {
        this.uri = uri;
    }

    @Override
    public String toString() {
        return "Response{" +
                "uri='" + uri + '\'' +
                '}';
    }
}
