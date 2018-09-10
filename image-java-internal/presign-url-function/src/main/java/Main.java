import java.io.OutputStreamWriter;
import java.net.HttpURLConnection;
import java.net.URL;

public class Main {
    public static void main(String[] args) {
        String uri = "https://stage-ringoid-origin-photo.s3.eu-west-1.amazonaws.com/f4664c6210abf40fe43d07e07418ef5606764669_photo.jpg?X-Amz-Security-Token=FQoGZXIvYXdzENn%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaDCH97AAzq1V%2FjJsm7yL5ARwIRd0PEe4HNwhY%2BF1KIjOnmN0lVy%2FI3Ax5FX%2FHN3gks%2FKnr5ZKTp0J0xB4tPGs1JiOkLXX89Qx6CU6I3QUeip9%2BhjRMvNPc971AyauLvirrxLaGaPdH7O6O9VnoIVPxiA3%2BGpsu%2B1JgyyIqHrG3KhsVuod3k%2F%2FkbpnZGRimJkOynN%2FRNltC%2Fl8nDONTHeiDUHSgZ2WMtiy7A52osAtoVTHUUg3eUkXb3hNJJeSSnd%2FFhDBz2SnFd7gALm47CXgYF3yYPYRxBBMHiSaK%2BQI923%2FltxAi0NCqshDTHtvpwTkA6QrQwlCKR5wzJ7fyf3wSqO5eREn6GyUuSiUyNjcBQ%3D%3D&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Date=20180910T075819Z&X-Amz-SignedHeaders=host&X-Amz-Expires=299&X-Amz-Credential=ASIAV7F6MSRDBVGK3EAK%2F20180910%2Feu-west-1%2Fs3%2Faws4_request&X-Amz-Signature=9cb63be3520b207e912ffe9ee74c0302156549ba75d35f973de7e08e1dbd0ae3";
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
