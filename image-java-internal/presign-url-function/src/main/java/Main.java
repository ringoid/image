import java.io.OutputStreamWriter;
import java.net.HttpURLConnection;
import java.net.URL;

public class Main {
    public static void main(String[] args) {
        String uri = "https://stage-ringoid-origin-photo.s3.eu-west-1.amazonaws.com/a4dcd33763007d4d99f69b2728e3c563aa62cf61_photo.jpg?X-Amz-Security-Token=FQoGZXIvYXdzEJj%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaDOPtG7nLWX5%2BjMbxGiL5AZ%2BE27NtG%2B1159UfgBGofoCeHfpmz5VzYmq8F3xMC2cLj0EAJhkau3S4fw5z14xDfO1HABApARn3ylM%2FyqTltquKi%2FIVVUdUQdYFC5fveNdOeMmMDLsWmgUqRKOl8L%2Fjz%2BU0PvPKWmtcvlPEO9swdg3k3ROtCSLZxfkK1bBqDj0dZ0MvbkCvKs%2B%2FzuhcHjEeG738hY%2FJfKHvunjQPsAiaXWro0MzR%2BrT11TqKZV%2FMyvu3%2FG4UOLXSqHPjqmgDjc47Kpx27kIEjZTTgGXJLqZnxbBMPuwPoe1Gceuhi%2F0V7gL35v8N4w%2B4GFkSz7ZvPACn0dz88X8JGVSOSjvm8rcBQ%3D%3D&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Date=20180907T144053Z&X-Amz-SignedHeaders=host&X-Amz-Expires=299&X-Amz-Credential=ASIAV7F6MSRDHMAZ33IZ%2F20180907%2Feu-west-1%2Fs3%2Faws4_request&X-Amz-Signature=36146f46b0e69e407afbc35f6317664b2786e69a2a82378b2161dc979abab1df";
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
