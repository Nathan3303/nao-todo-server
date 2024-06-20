const mongoose = require("mongoose");

module.exports = async function connectMongo(
    hostAndPort,
    dbName,
    successCallback
) {
    if (typeof successCallback !== "function") successCallback = () => {};
    mongoose.connection.once("open", successCallback);
    mongoose.connection.on(
        "error",
        console.error.bind(console, "MongoDB connection error:")
    );
    mongoose.connection.once("close", () => {
        console.log("MongoDB connection closed");
    });
    mongoose.set("strictQuery", true);
    return await mongoose.connect(`mongodb://${hostAndPort}/${dbName}`);
};
