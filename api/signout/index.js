const { Session } = require("../../models");
const { checkMethod, buildRD } = require("../../utils");

module.exports = async (request, response) => {
    if (checkMethod(request, response, "DELETE")) return;
    const { jwt } = request.query;
    if (!jwt) {
        response.status(200).json(buildRD.error("Invalid queries"));
        return;
    }
    try {
        const jwtPayload = jwt.split(".")[1];
        const user = JSON.parse(atob(jwtPayload));
        const deleteRes = await Session.deleteOne({
            token: jwt,
            userId: user.id,
        });
        if (deleteRes.deletedCount) {
            response.status(200).json(buildRD.success("Sign out succesfully"));
            return;
        }
    } catch (error) {
        response.statuc(200).json(buildRD.error(error.message));
    }
};
