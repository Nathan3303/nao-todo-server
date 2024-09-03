const moment = require("moment");
const { User, Session } = require("../../models");
const {
    buildRD,
    checkMethod,
    checkQueryLength,
    toSign,
    makeJWT,
} = require("../../utils");

const { ObjectId } = require("mongodb");

module.exports = async (request, response) => {
    if (checkMethod(request, response, "GET")) return;
    if (checkQueryLength(request, response)) return;
    const { jwt } = request.query;
    const [jwtHeader, jwtPayload, jwtSignature] = jwt.split(".");
    if (jwtSignature !== toSign(jwtHeader, jwtPayload)) {
        response.status(200).json(buildRD.error("Token is invalid."));
        return;
    }
    try {
        const session = await Session.findOne({ token: jwt });
        if (!session) {
            response.status(200).json(buildRD.error("Session not found."));
            return;
        }
        if (moment().isAfter(session.expiresAt)) {
            const result = await Session.deleteOne({ _id: session._id });
            console.log(result);
            response.status(200).json(buildRD.error("Session expired."));
            return;
        }
        const user = await User.findOne({
            _id: new ObjectId(session.userId),
        }).select({
            _id: 0,
            id: { $toString: "$_id" },
            email: 1,
            nickName: 1,
        });
        if (!user) {
            response.status(200).json(buildRD.error("User not found."));
            return;
        }
        const newJWT = makeJWT(user);
        const updateRes = await Session.updateOne(
            { _id: session._id },
            { token: newJWT.token }
        );
        response.status(200).json(buildRD.success(newJWT.token));
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
