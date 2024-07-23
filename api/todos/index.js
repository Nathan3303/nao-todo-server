const buildRD = require("../../utils/build-response-data");
const { verifyJWT } = require("../../utils/make-jwt");

const getTodos = require("./get-todos");
const pipelines = require("./pipelines");

module.exports = async (request, response) => {
    // Check authorization
    const { authorization } = request.headers;
    if (!authorization) {
        response
            .status(200)
            .json(buildRD.error("Authorization header is missing."));
        return;
    }
    const token = authorization.replace("Bearer ", "");
    const verifyResult = verifyJWT(token);
    if (!verifyResult) {
        response.status(200).json(buildRD.error("Invalid token."));
        return;
    }

    // Distribute request to corresponding method
    switch (request.method) {
        case "GET":
            await getTodos(request, response, pipelines);
            break;
        default:
            response.status(200).json(buildRD.error("Invalid method."));
            break;
    }
};
