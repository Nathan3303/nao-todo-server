import {
    getModelForClass,
    prop,
    modelOptions,
    index,
    mongoose
} from '@typegoose/typegoose';

@index({ userId: 1, todoId: 1 })
@index({ userId: 1, todoId: 1, title: 1 })
@modelOptions({ schemaOptions: { timestamps: true, collection: 'events' } })
class Event {
    @prop({ required: true, ref: 'User' })
    userId: mongoose.Types.ObjectId;

    @prop({ required: true, ref: 'Todo' })
    todoId: mongoose.Types.ObjectId;

    @prop({ required: true })
    title: string;

    @prop({ default: '' })
    description: string;

    @prop({ default: false })
    isDone: boolean;

    @prop({ default: null })
    doneAt: Date;

    @prop({ default: false })
    isTopped: boolean;
}

export default getModelForClass(Event);
export { Event };
