import { Project } from '@nao-todo-server/models';
import { ObjectId, parseToBool } from '@nao-todo-server/utils';

const handleUserId = (userId: string) => {
    return userId ? Project.aggregate().match({ userId: new ObjectId(userId) }).pipeline() : [];
};

const handleTitle = (title: string) => {
    return title
        ? Project.aggregate()
              .match({ title: { $regex: title, $options: 'i' } })
              .pipeline()
        : [];
};

const handleIsDeleted = (isDeleted: boolean) => {
    if (isDeleted === null || isDeleted === undefined) {
        return [];
    }
    const _bool = parseToBool(isDeleted);
    return Project.aggregate().match({ isDeleted: _bool }).pipeline();
};

const handleIsFinished = (isFinished: boolean) => {
    if (isFinished === null || isFinished === undefined) {
        return [];
    }
    const _bool = parseToBool(isFinished);
    return Project.aggregate().match({ isFinished: _bool }).pipeline();
};

const handleIsArchived = (isArchived: boolean) => {
    if (isArchived === null || isArchived === undefined) {
        return [];
    }
    const _bool = parseToBool(isArchived);
    return Project.aggregate().match({ isArchived: _bool }).pipeline();
};

const handlePage = (page: string, limit: string) => {
    const _page = parseInt(page) || 1;
    const _limit = parseInt(limit) || 10;
    return Project.aggregate()
        .skip((_page - 1) * _limit)
        .limit(_limit)
        .pipeline();
};

const handleSort = (sort: string) => {
    if (sort === null || sort === undefined) {
        return [];
    }
    const sortObject: Record<string, any> = {};
    const splited = sort.split(',');
    splited.forEach(s => {
        if (s.startsWith('-')) {
            s = s.slice(1);
            sortObject[s] = -1;
        } else {
            sortObject[s] = 1;
        }
    });
    return Project.aggregate().sort(sortObject).pipeline();
};

const handleOutput = () => {
    return Project.aggregate()
        .project({
            _id: 0,
            id: { $toString: '$_id' },
            userId: 1,
            title: 1,
            description: 1,
            state: 1,
            isArchived: 1,
            isFavorite: 1,
            isDeleted: 1,
            isFinished: 1,
            createdAt: 1,
            updatedAt: 1,
            deletedAt: 1,
            finishedAt: 1,
            archivedAt: 1,
            preference: 1
        })
        .pipeline();
};

export default {
    handleUserId,
    handleTitle,
    handleIsDeleted,
    handleIsFinished,
    handleIsArchived,
    handlePage,
    handleSort,
    handleOutput
};
