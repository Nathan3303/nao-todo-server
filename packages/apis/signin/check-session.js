const moment = require("moment");
const { Session } = require("../../models");
const { makeJWT } = require("../../utils");

module.exports = async function checkSession(request, user) {
    let session = await Session.findOne({ userId: user.id });
    const createdAt = moment().utcOffset(8).utc().format();
    const expiresAt = moment().add(7, "days").utcOffset(8).utc().format();
    if (session) {
        await Session.updateOne(
            { _id: session._id },
            { $set: { createdAt, expiresAt } }
        );
        return session.token;
    } else {
        const jwt = makeJWT(user);
        session = await Session.create({
            userId: user.id,
            token: jwt.token,
            createdAt,
            expiresAt,
            browser: request.headers["user-agent"],
        });
        return session ? jwt.token : null;
    }
};
