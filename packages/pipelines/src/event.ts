import { Event } from '@nao-todo-server/models';
import { ObjectId, parseToBool } from '@nao-todo-server/utils';

const handleUserId = (userId: string) => {
    return userId
        ? Event.aggregate()
              .match({ userId: new ObjectId(userId) })
              .pipeline()
        : [];
};

const handleTodoId = (todoId?: string) => {
    return todoId
        ? Event.aggregate()
              .match({ todoId: new ObjectId(todoId) })
              .pipeline()
        : [];
};

const handleId = (id?: string) => {
    return id
        ? Event.aggregate()
              .match({ _id: new ObjectId(id) })
              .pipeline()
        : [];
};

const handleTitle = (title?: string) => {
    return title
        ? Event.aggregate()
              .match({ title: { $regex: title, $options: 'i' } })
              .pipeline()
        : [];
};

const handleIsDone = (isDone?: boolean) => {
    return isDone
        ? Event.aggregate()
              .match({ isDone: parseToBool(isDone) })
              .pipeline()
        : [];
};

const handleIsTopped = (isTopped?: boolean) => {
    return isTopped
        ? Event.aggregate()
              .match({ isTopped: parseToBool(isTopped) })
              .pipeline()
        : [];
};

const handlePage = (page: number, limit: number) => {
    const _page = page || 1;
    const _limit = limit || 10;
    return Event.aggregate()
        .skip((_page - 1) * _limit)
        .limit(_limit)
        .pipeline();
};

const handleSelectFields = (
    fieldsOptions?: Record<string, any> | undefined
) => {
    fieldsOptions = fieldsOptions || {
        _id: 0,
        id: { $toString: '$_id' },
        title: 1,
        isDone: 1,
        isTopped: 1
    };
    return Event.aggregate().project(fieldsOptions).pipeline();
};

export default {
    handleUserId,
    handleTodoId,
    handleId,
    handleTitle,
    handleIsDone,
    handleIsTopped,
    handlePage,
    handleSelectFields
};
