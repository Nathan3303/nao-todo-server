import { Tag } from '@nao-todo-server/models';
import { ObjectId, parseToBool } from '@nao-todo-server/utils';

const handleUserId = (userId: string) => {
    return userId ? Tag.aggregate().match({ userId: new ObjectId(userId) }).pipeline() : [];
};

const handleName = (name?: string) => {
    return name
        ? Tag.aggregate()
              .match({ name: { $regex: `.*${name}.*`, $options: 'i' } })
              .pipeline()
        : [];
};

const handleIsDeleted = (isDeleted?: boolean) => {
    return isDeleted
        ? Tag.aggregate()
              .match({ isDeleted: parseToBool(isDeleted) })
              .pipeline()
        : [];
};

const handlePage = (page: string, limit: string) => {
    const _page = parseInt(page) || 1;
    const _limit = parseInt(limit) || 10;
    return Tag.aggregate()
        .skip((_page - 1) * _limit)
        .limit(_limit)
        .pipeline();
};

const handleSelectFields = (fieldsOptions?: Record<string, any>) => {
    fieldsOptions = fieldsOptions || {
        _id: 0,
        id: { $toString: '$_id' },
        name: 1,
        color: 1,
        createdAt: 1,
        updatedAt: 1
    };
    return Tag.aggregate().project(fieldsOptions).pipeline();
};

export default {
    handleUserId,
    handleName,
    handleIsDeleted,
    handlePage,
    handleSelectFields
};
