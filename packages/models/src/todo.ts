import {
    getModelForClass,
    prop,
    modelOptions,
    index
} from '@typegoose/typegoose';

@modelOptions({ schemaOptions: { _id: false } })
class TodoDueDate {
    startAt: Date;
    endAt: Date;
}

@index({ userId: 1, projectId: 1 })
@index({ userId: 1, projectId: 1, name: 1, state: 1, priority: 1 })
@modelOptions({
    schemaOptions: { timestamps: true, collection: 'todos' },
    options: { allowMixed: 0 }
})
class Todo {
    @prop({ required: true })
    userId: string;

    @prop({ required: true })
    projectId: string;

    @prop({ required: true })
    name: string;

    @prop({ default: '' })
    description: string;

    @prop({ default: 0, enum: [0, 1, 2] })
    state: number;

    @prop({ default: 0, enum: [0, 1, 2, 3] })
    priority: number;

    @prop({ default: [] })
    tags: string[];

    // @prop({ default: false })
    // isDone: boolean;

    @prop({ default: null })
    dueDate: TodoDueDate;

    @prop({ default: false })
    isDeleted: boolean;

    @prop({ default: false })
    isArchived: boolean;

    @prop({ default: false })
    isFavorited: boolean;
}

export default getModelForClass(Todo);
export { Todo };
