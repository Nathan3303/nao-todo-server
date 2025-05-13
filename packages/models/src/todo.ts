import {
    getModelForClass,
    index,
    modelOptions,
    mongoose,
    prop
} from '@typegoose/typegoose';

@modelOptions({ schemaOptions: { _id: false } })
class TodoDueDate {
    @prop({ default: null })
    startAt: Date;

    @prop({ default: null })
    endAt: Date;
}

@index({ userId: 1, projectId: 1 })
@index({ userId: 1, projectId: 1, name: 1, state: 1, priority: 1 })
@modelOptions({
    schemaOptions: { timestamps: true, collection: 'todos' },
    options: { allowMixed: 0 }
})
class Todo {
    @prop({ required: true, ref: 'User' })
    userId: mongoose.Types.ObjectId;

    @prop({ required: true, ref: 'Project' })
    projectId: mongoose.Types.ObjectId;

    @prop({ required: true })
    name: string;

    @prop({ default: '' })
    description: string;

    @prop({ default: 'todo', enum: ['todo', 'in-progress', 'done'] })
    state: string;

    @prop({ default: 'low', enum: ['low', 'medium', 'high', 'urgent'] })
    priority: string;

    @prop({ default: [mongoose.Schema.Types.ObjectId], ref: 'Tag' })
    tags: mongoose.Types.ObjectId[];

    @prop({ default: { startAt: null, endAt: null } })
    dueDate: TodoDueDate;

    @prop({ default: false })
    isDeleted: boolean;

    @prop({ default: null })
    deletedAt: Date;

    @prop({ default: false })
    isArchived: boolean;

    @prop({ default: false })
    isFavorited: boolean;

    @prop({ default: false })
    isGivenUp: boolean;
}

export default getModelForClass(Todo);
export { Todo };
