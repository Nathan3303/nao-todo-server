const buildRD = require("../../utils/build-response-data");
const getProjects = require("./get-project");
const createProject = require("./create-project");
const deleteProject = require("./delete-project");
const updateProject = require("./update-project");

module.exports = async (request, response) => {
    switch (request.method) {
        case "GET":
            await getProjects(request, response);
            break;
        case "POST":
            await createProject(request, response);
            break;
        case "DELETE":
            await deleteProject(request, response);
            break;
        case "PUT":
            await updateProject(request, response);
            break;
        default:
            response.status(200).json(buildRD.error("Method not allowed"));
            break;
    }
};
