import {
    getModelForClass,
    prop,
    modelOptions,
    index
} from '@typegoose/typegoose';

@index({ userId: 1, todoId: 1 })
@index({ userId: 1, todoId: 1, title: 1 })
@modelOptions({ schemaOptions: { timestamps: true, collection: 'events' } })
class Event {
    @prop({ required: true })
    userId: string;

    @prop({ required: true })
    todoId: string;

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
