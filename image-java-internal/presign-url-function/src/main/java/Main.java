import java.io.OutputStreamWriter;
import java.net.HttpURLConnection;
import java.net.URL;

public class Main {
    public static void main(String[] args) {
        String uri = "https://stage-ringoid-origin-photo.s3.eu-west-1.amazonaws.com/668a6a4d1c201718c194d2a7503d812934ee509a_photo.jpg?X-Amz-Security-Token=FQoGZXIvYXdzENv%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaDDAaI%2FZIWgZk6VQJISL5AdyZmNCuiiPkAuBIiduCjevOB1eJA38i7VAqlsm5Q%2BE7RbQXha4mjHmG7v70ocXB21EyEeXnx2Z8EUpK6sfCVR9LeU%2FwSMU%2F0pFsrSzlDxUP1VU%2B%2BbuK9Q3dbFsyC0w%2BfYmStQALJjV6Sjuj6h5sxEW7kczGCYwfjRIn%2FnaYz5Ieltb6wVb8gnnPaiM4baeSFMfGFtkDbdnCTrh5drwKzo2HA4uCH6ABxMBtndgjnqdmBwYWKA4nTVFP6rIo6745tsGYgf6lqpP4BufX5dmv2UA4z99OgsVdw0MMqSou2CTxqBS7NbSMBc9XQBoTYpAhbwkBCem%2F6evZiijOgdncBQ%3D%3D&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Date=20180910T100353Z&X-Amz-SignedHeaders=host&X-Amz-Expires=300&X-Amz-Credential=ASIAV7F6MSRDPCSRHDGG%2F20180910%2Feu-west-1%2Fs3%2Faws4_request&X-Amz-Signature=c3ced1ddb593472b109015aaa5b1533e6af78a15cb29b0fd3a8c33659b1d7ef6";
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
