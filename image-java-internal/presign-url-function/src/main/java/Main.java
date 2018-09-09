import java.io.OutputStreamWriter;
import java.net.HttpURLConnection;
import java.net.URL;

public class Main {
    public static void main(String[] args) {
        String uri = "https://stage-ringoid-origin-photo.s3.eu-west-1.amazonaws.com/e0be884d6691eeeb425c55a0f88250c2a9306925_photo.jpg?X-Amz-Security-Token=FQoGZXIvYXdzEMf%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaDOUY%2FQe%2BsoKJoC722SL5ARH%2FCchfjRhNLBHq7Yo8Jl2TsFIyT0DMEkOPOMawF8OSLXw4reE5MyeJcJbwK%2BvR5%2Fa4vE5PD5l%2F%2FmOiS1Oga3PrFhFqKY45HzveeZHTr%2BCYxObLEd1kR4HXbiwelbY%2FLRlVo8MElWFu%2Bno05Om6YJEGe0XNnhJ6FnYPUK5uMv8gmK3UkRn9YPV8C7mwt1rqjxmEK8Yy15bKb589SNCIaDl20i9LqopVadWgKdA0%2F31yPOTjVPMmyCnrrv6ZY5RzEeAIzWyo3%2F2r5Y75AEc7yx%2Frjqql0LmyBm7tSm4zJ6qny1Hp0Q9Y3s4XjrHleLVep3uxwKRzuzvHxii9utTcBQ%3D%3D&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Date=20180909T131653Z&X-Amz-SignedHeaders=host&X-Amz-Expires=299&X-Amz-Credential=ASIAV7F6MSRDJHDLHZPK%2F20180909%2Feu-west-1%2Fs3%2Faws4_request&X-Amz-Signature=5ba1a7c7cf5b134bbada0fa4bc01ee0299ede1a8b44ee32dcafc4d0d9579a152";
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
