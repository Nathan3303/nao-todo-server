import { Event } from '@nao-todo-server/models';
import { ObjectId, parseToBool } from '@nao-todo-server/utils';

export const handleUserId = (userId: string) => {
    return userId
        ? Event.aggregate()
              .match({ userId: new ObjectId(userId) })
              .pipeline()
        : [];
};

export const handleTodoId = (todoId: string) => {
    return todoId
        ? Event.aggregate()
              .match({ todoId: new ObjectId(todoId) })
              .pipeline()
        : [];
};

export const handleId = (id: string) => {
    return id
        ? Event.aggregate()
              .match({ _id: new ObjectId(id) })
              .pipeline()
        : [];
};

export const handleTitle = (title: string) => {
    return title
        ? Event.aggregate()
              .match({ title: { $regex: title, $options: 'i' } })
              .pipeline()
        : [];
};

export const handleIsDone = (isDone: string) => {
    return isDone
        ? Event.aggregate()
              .match({ isDone: parseToBool(isDone) })
              .pipeline()
        : [];
};

export const handleIsTopped = (isTopped: string) => {
    return isTopped
        ? Event.aggregate()
              .match({ isTopped: parseToBool(isTopped) })
              .pipeline()
        : [];
};

export const handlePage = (page: string, limit: string) => {
    const _page = parseInt(page) || 1;
    const _limit = parseInt(limit) || 10;
    return Event.aggregate()
        .skip((_page - 1) * _limit)
        .limit(_limit)
        .pipeline();
};

export const handleSelectFields = (fieldsOptions: Record<string, any> | undefined) => {
    fieldsOptions = fieldsOptions || {
        _id: 0,
        id: { $toString: '$_id' },
        title: 1,
        isDone: 1,
        isTopped: 1
    };
    return Event.aggregate().project(fieldsOptions).pipeline();
};
