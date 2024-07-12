const express = require("express");
const mongoose = require("mongoose");
const cors = require("cors");

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
    .connect(`mongodb://localhost/naotodo`)
    .then(() => {
        app.listen(3002, () => {
            console.log("NaoTodoServer(dev) is running on port 3002");
        });
    });
