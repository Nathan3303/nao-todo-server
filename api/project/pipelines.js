const Project = require("../../models/project");
const ObjectId = require("mongoose").Types.ObjectId;

const matchProjectById = (id, isGetRaw = true) => {
    id = id || "";
    const aggregate = Project.aggregate()
        .match({ _id: new ObjectId(id) })
        .project({
            _id: 0,
            id: { $toString: "$_id" },
            name: 1,
            description: 1,
            owner: 1,
            created_at: 1,
            updated_at: 1,
        })
        .allowDiskUse(true);
    return isGetRaw ? aggregate.pipeline() : aggregate;
};

const matchProjectByNameAndOwner = (name, owner, isGetRaw = true) => {
    name = name || "";
    owner = owner || "";
    const aggregate = Project.aggregate()
        .match({
            name: { $regex: name, $options: "i" },
            owner: new ObjectId(owner),
        })
        .project({
            _id: 0,
            id: { $toString: "$_id" },
            name: 1,
            description: 1,
            owner: 1,
            created_at: 1,
            updated_at: 1,
        })
        .allowDiskUse(true);
    return isGetRaw ? aggregate.pipeline() : aggregate;
};

module.exports = {
    matchProjectById,
    matchProjectByNameAndOwner,
};
