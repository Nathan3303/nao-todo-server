const buildRD = require("./build-response-data");

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

module.exports = { checkMethod, checkQueryLength };
