const buildRD = require("./build-response-data");
const { toSign, makeJWT } = require("./make-jwt");
const serialExecution = require("./serial-execution");

function checkMethod(request, response, method) {
    if (request.method !== method) {
        console.log("Method not allowed");
        response.status(405).json(buildRD.error("Method not allowed"));
        return 1;
    }
    return 0;
}

function checkQueryLength(request, response) {
    if (Object.keys(request.query).length === 0) {
        response
            .status(400)
            .json(buildRD.error("No query parameters provided"));
        return 1;
    }
    return 0;
}

const makeBoolean = (value) => {
    if (typeof value === "string") {
        return value.toLowerCase() === "true";
    } else if (typeof value === "boolean") {
        return value;
    } else if (typeof value === "number") {
        return value === 1;
    } else {
        return false;
    }
};

module.exports = {
    buildRD,
    toSign,
    makeJWT,
    serialExecution,
    checkMethod,
    checkQueryLength,
    makeBoolean,
};
