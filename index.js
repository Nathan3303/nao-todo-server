const express = require("express");
const mongoose = require("mongoose");
const cors = require("cors");
const https = require("https");
const fs = require("fs");
const path = require("path");

const app = express();

app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.get("/api/projects", require("./api/projects"));
app.get("/api/todos", require("./api/todos"));
app.get("/api/analysis", require("./api/analysis"));

app.use("/api/project", require("./api/project"));
app.use("/api/todo", require("./api/todo"));
app.use("/api/event", require("./api/event"));

app.use("/", (_, res) => res.end("Hello World!"));

mongoose
    .set("strictQuery", true)
    .connect(`mongodb://127.0.0.1/naotodo`)
    .then(() => {
        https
            .createServer(
                {
                    key: fs.readFileSync(
                        path.join(__dirname, "ssl/privkey.pem")
                    ),
                    cert: fs.readFileSync(
                        path.join(__dirname, "ssl/fullchain.pem")
                    ),
                },
                app
            )
            .listen(3002, () => {
                console.log("NaoTodoServer(prod) is running on port 3002");
            });
    });
