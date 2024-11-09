import { Todo } from '@nao-todo-server/models';
import { ObjectId, parseToBool } from '@nao-todo-server/utils';

const handleUserId = (userId?: string) => {
    return userId ? Todo.aggregate().match({ userId }).pipeline() : [];
};

const handleProjectId = (projectId?: string) => {
    return projectId ? Todo.aggregate().match({ projectId }).pipeline() : [];
};

const handleTagId = (tagId: string) => {
    return tagId
        ? Todo.aggregate()
              .match({ tags: { $in: [tagId] } })
              .pipeline()
        : [];
};

const handleId = (id: string) => {
    return id
        ? Todo.aggregate()
              .match({ _id: new ObjectId(id) })
              .pipeline()
        : [];
};

const handleName = (name?: string) => {
    return name
        ? Todo.aggregate()
              .match({ name: { $regex: name, $options: 'i' } })
              .pipeline()
        : [];
};

const handleState = (state?: string) => {
    if (!state) return [];
    const stateStr = ['todo', 'doing', 'done'];
    let stateNums: number[] = [];
    const splitedState = state.split(',');
    for (const s of splitedState) {
        const stateNumber = stateStr.indexOf(s);
        if (stateNumber === -1) continue;
        stateNums.push(stateNumber);
    }
    return Todo.aggregate()
        .match({ state: { $in: stateNums } })
        .pipeline();
};

const handlePriority = (priority?: string) => {
    if (!priority) return [];
    const splitedPriority = priority.split(',');
    return Todo.aggregate()
        .match({ priority: { $in: splitedPriority } })
        .pipeline();
};

const handleIsFavorited = (isFavorited?: boolean) => {
    return isFavorited
        ? Todo.aggregate()
              .match({ isPinned: parseToBool(isFavorited) })
              .pipeline()
        : [];
};

const handleIsDeleted = (isDeleted?: boolean) => {
    return isDeleted
        ? Todo.aggregate()
              .match({ isDeleted: parseToBool(isDeleted) })
              .pipeline()
        : [];
};

const handlePage = (page: number, limit: number) => {
    const _page = page || 1;
    const _limit = limit || 10;
    return Todo.aggregate()
        .skip((_page - 1) * _limit)
        .limit(_limit)
        .pipeline();
};

const handleSort = (sort?: string) => {
    if (!sort) {
        return Todo.aggregate().sort({ createdAt: 'desc' }).pipeline();
    }
    let [field, order] = sort.split(':');
    if (!field) return [];
    if (field === 'endAt') field = 'dueDate.endAt';
    else if (field === 'startAt') field = 'dueDate.startAt';
    order = order || 'asc';
    let _sortOption: Record<string, any> = {};
    _sortOption[field] = order;
    return Todo.aggregate().sort(_sortOption).pipeline();
};

const handleLookupProject = () => {
    return Todo.aggregate()
        .lookup({
            from: 'projects',
            localField: 'projectId',
            foreignField: '_id',
            as: '_project'
        })
        .unwind({ path: '$_project', preserveNullAndEmptyArrays: true })
        .pipeline();
};

const handleLookupTags = () => {
    return Todo.aggregate()
        .lookup({
            from: 'tags',
            localField: 'tags',
            foreignField: '_id',
            as: 'tagsInfo'
        })
        .pipeline();
};

const handleSelectFields = (fieldsOptions?: Record<string, any>) => {
    fieldsOptions = fieldsOptions || {
        _id: 0,
        id: { $toString: '$_id' },
        projectId: 1,
        project: {
            title: '$_project.title'
        },
        name: 1,
        state: 1,
        priority: 1,
        tags: 1,
        tagsInfo: {
            $map: {
                input: '$tagsInfo',
                as: 'tag',
                in: {
                    id: { $toString: '$$tag._id' },
                    name: '$$tag.name',
                    color: '$$tag.color'
                }
            }
        },
        isDone: 1,
        createdAt: 1,
        updatedAt: 1,
        description: 1,
        isPinned: 1,
        dueDate: 1,
        isDeleted: 1
    };
    return Todo.aggregate().project(fieldsOptions).pipeline();
};

const handleGroupByState = () => {
    return Todo.aggregate()
        .group({
            _id: '$state',
            total: { $sum: 1 },
            totalTodos: { $sum: 1 }
        })
        .pipeline();
};

const handleGroupByPriority = () => {
    return Todo.aggregate()
        .group({
            _id: '$priority',
            total: { $sum: 1 },
            totalTodos: { $sum: 1 }
        })
        .pipeline();
};

const handleCountTotal = () => {
    return Todo.aggregate().count('total').pipeline();
};

const handleRelativeDate = (relativeDate?: string) => {
    if (!relativeDate) return [];
    let agg = null;
    switch (relativeDate) {
        case 'today':
            agg = Todo.aggregate().match({
                'dueDate.endAt': {
                    $gte: new Date(new Date().setHours(0, 0, 0, 0)),
                    $lte: new Date(new Date().setHours(23, 59, 59, 999))
                }
            });
            break;
        case 'tomorrow':
            agg = Todo.aggregate().match({
                'dueDate.endAt': {
                    $gte: new Date(
                        new Date().setHours(0, 0, 0, 0) + 24 * 60 * 60 * 1000
                    ),
                    $lte: new Date(
                        new Date().setHours(23, 59, 59, 999) +
                            24 * 60 * 60 * 1000
                    )
                }
            });
            break;
        case 'week':
            agg = Todo.aggregate().match({
                'dueDate.endAt': {
                    $gte: new Date(new Date().setHours(0, 0, 0, 0)),
                    $lte: new Date(
                        new Date().setHours(23, 59, 59, 999) +
                            7 * 24 * 60 * 60 * 1000
                    )
                }
            });
            break;
        default:
            agg = null;
    }
    return agg ? agg.sort({ 'dueDate.endAt': 1 }).pipeline() : [];
};

export default {
    handleUserId,
    handleProjectId,
    handleName,
    handlePriority,
    handleRelativeDate,
    handleState,
    handleTagId,
    handleId,
    handleIsFavorited,
    handleIsDeleted,
    handlePage,
    handleGroupByPriority,
    handleCountTotal,
    handleLookupProject,
    handleLookupTags,
    handleSelectFields,
    handleGroupByState,
    handleSort
};
