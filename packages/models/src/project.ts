import mongoose from 'mongoose';

const ObjectId = mongoose.Schema.Types.ObjectId;

const projectSchema = new mongoose.Schema({
    /**
     * The user who created the project
     */
    userId: { type: ObjectId, ref: 'User' },
    /**
     * The title of the project
     */
    title: { type: String, required: true },
    /**
     * The description of the project
     */
    description: { type: String, default: null },
    /**
     * The date when the project was created
     * 0 - Not started
     * 1 - In progress
     * 2 - Done
     */
    state: { type: [0, 1, 2], default: 0 },
    /**
     * The date when the project was archived
     */
    isArchived: { type: Boolean, default: false },
    /**
     * The date when the project was deleted
     */
    isDeleted: { type: Boolean, default: false },
    /**
     * The date when the project was finished
     */
    isFinished: { type: Boolean, default: false },
    /**
     * The date when the project was favorited
     */
    isFavorite: { type: Boolean, default: false },
    /**
     * The date when the project was created
     */
    createdAt: { type: Date, default: Date.now },
    /**
     * The date when the project was updated
     */
    updatedAt: { type: Date, default: Date.now },
    /**
     * The date when the project was deleted
     */
    deletedAt: { type: Date, default: null },
    /**
     * The date when the project was finished
     */
    finishedAt: { type: Date, default: null },
    /**
     * The date when the project was favorited
     */
    archivedAt: { type: Date, default: null },
    /**
     * The user's preference for the project
     */
    preference: {
        viewType: { type: String, default: 'table' },
        filterInfo: {
            type: Object,
            default: {
                isDeleted: false
            }
        },
        sortInfo: {
            type: { field: String, order: String },
            default: {
                field: 'createdAt',
                order: 'desc'
            }
        },
        columns: {
            type: Object,
            default: {
                state: false,
                priority: true,
                project: true,
                description: true,
                endAt: true,
                createdAt: false,
                updatedAt: false
            }
        }
    }
});

export default mongoose.model('Project', projectSchema);
