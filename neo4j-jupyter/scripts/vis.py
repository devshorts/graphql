import json
import uuid
from IPython.display import IFrame


def draw(cipher, options={}):
    html = """
<html>
<head>
    <script src="https://rawgit.com/neo4j-contrib/neovis.js/master/dist/neovis.js"></script>
    <script type="text/javascript">

        var viz;

        function draw() {{
            var config = {{
                container_id: "viz",
                server_url: "bolt://localhost:7687",
                server_user: "",
                server_password: "",
                labels: {options},
                initial_cypher: "{cipher}"
            }};

            viz = new NeoVis.default(config);
            viz.render();
        }}
    </script>
</head>
<body onload="draw()">
<div id="viz"></div>
</body>
</html>
    """

    unique_id = str(uuid.uuid4())

    sanitized_cipher =cipher.replace("\"", "'").replace("\n", " ")

    html = html.format(cipher=sanitized_cipher, options=json.dumps(options))

    filename = "figure/graph-{}.html".format(unique_id)

    file = open(filename, "w")
    file.write(html)
    file.close()

    return IFrame(filename, width="100%", height="400")
