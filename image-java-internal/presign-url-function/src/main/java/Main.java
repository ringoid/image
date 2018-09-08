import java.io.OutputStreamWriter;
import java.net.HttpURLConnection;
import java.net.URL;

public class Main {
    public static void main(String[] args) {
        String uri = "https://stage-ringoid-origin-photo.s3.eu-west-1.amazonaws.com/3ca6eb354884de7be74e4bc5ecb2c5479a2198c6_photo.jpg?X-Amz-Security-Token=FQoGZXIvYXdzEJr%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaDJN3%2FMS2yZoJLG6ZxSL5AXMndcqPshadci08k%2Brwnf8zjaOvZ3FwFaxKR1GlQ3pArNSw9Zxs%2BvHTrWdKlemEAmMBRZoZMRMlUoNHI95cu7PT4vM5s479AUeyX7WEVHdjvh8dBuIjj0U9FVQBHdYfBAx0todv75%2BYqXDFDc4Z18p0f3QogqeAKkG47tr5J86H8EcvSy8WZLoYFXUP76rVfUnUtkrnXu%2FD7ZvAc2GdMXWArSq0%2Bit3V%2FbOBhHn2SnWNtdLuKvnEqvCRmtuGGMtHxz0QvrxqiuRh579QHnH4rUBq0wtEmIJbwulTi57W%2Bs7%2FwngtQTXt7Ct%2F5c4aF%2BCnMzppjfQIkIV%2BSjP1crcBQ%3D%3D&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Date=20180907T164359Z&X-Amz-SignedHeaders=host&X-Amz-Expires=300&X-Amz-Credential=ASIAV7F6MSRDMTP24AUQ%2F20180907%2Feu-west-1%2Fs3%2Faws4_request&X-Amz-Signature=7ffb5aeb0fdd9c5ecd345fb443929e0d8eff174f27959cc64547c4c11afd3846";
        try {
            URL nUrl = new URL(uri);
            HttpURLConnection connection = (HttpURLConnection) nUrl.openConnection();
            connection.setDoOutput(true);
            connection.setRequestMethod("PUT");
            OutputStreamWriter out = new OutputStreamWriter(connection.getOutputStream());
            out.write("This text uploaded as an object via presigned URL from the local PC");
            out.close();

            connection.getResponseCode();
            System.out.println("HTTP response code: " + connection.getResponseCode());
        } catch (Throwable t) {
            t.printStackTrace();
        }
    }
}
