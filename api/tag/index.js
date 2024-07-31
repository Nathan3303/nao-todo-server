const pipelines = require("../../pipelines/project");
const buildRD = require("../../utils/build-response-data");
const createTag = require("./create-tag");
const updateTag = require("./update-tag");

module.exports = async (request, response) => {
    const { method } = request;

    switch (method) {
        case "POST":
            await createTag(request, response, pipelines);
            break;
        case "PUT":
            await updateTag(request, response, pipelines);
            break;
        default:
            response.status(200).json(buildRD.error("Invalid method."));
            break;
    }
};
