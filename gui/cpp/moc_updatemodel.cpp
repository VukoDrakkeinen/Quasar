/****************************************************************************
** Meta object code from reading C++ file 'updatemodel.h'
**
** Created by: The Qt Meta Object Compiler version 67 (Qt 5.5.0)
**
** WARNING! All changes made in this file will be lost!
*****************************************************************************/

#include "updatemodel.h"
#include <QtCore/qbytearray.h>
#include <QtCore/qmetatype.h>
#if !defined(Q_MOC_OUTPUT_REVISION)
#error "The header file 'updatemodel.h' doesn't include <QObject>."
#elif Q_MOC_OUTPUT_REVISION != 67
#error "This file was generated using the moc from 5.5.0. It"
#error "cannot be used with the include files from this version of Qt."
#error "(The moc has changed too much.)"
#endif

QT_BEGIN_MOC_NAMESPACE
struct qt_meta_stringdata_UpdateStatus_t {
    QByteArrayData data[6];
    char stringdata0[55];
};
#define QT_MOC_LITERAL(idx, ofs, len) \
    Q_STATIC_BYTE_ARRAY_DATA_HEADER_INITIALIZER_WITH_OFFSET(len, \
    qptrdiff(offsetof(qt_meta_stringdata_UpdateStatus_t, stringdata0) + ofs \
        - idx * sizeof(QByteArrayData)) \
    )
static const qt_meta_stringdata_UpdateStatus_t qt_meta_stringdata_UpdateStatus = {
    {
QT_MOC_LITERAL(0, 0, 12), // "UpdateStatus"
QT_MOC_LITERAL(1, 13, 4), // "Enum"
QT_MOC_LITERAL(2, 18, 9), // "NoUpdates"
QT_MOC_LITERAL(3, 28, 8), // "Updating"
QT_MOC_LITERAL(4, 37, 11), // "NewChapters"
QT_MOC_LITERAL(5, 49, 5) // "Error"

    },
    "UpdateStatus\0Enum\0NoUpdates\0Updating\0"
    "NewChapters\0Error"
};
#undef QT_MOC_LITERAL

static const uint qt_meta_data_UpdateStatus[] = {

 // content:
       7,       // revision
       0,       // classname
       0,    0, // classinfo
       0,    0, // methods
       0,    0, // properties
       1,   14, // enums/sets
       0,    0, // constructors
       0,       // flags
       0,       // signalCount

 // enums: name, flags, count, data
       1, 0x0,    4,   18,

 // enum data: key, value
       2, uint(UpdateStatus::NoUpdates),
       3, uint(UpdateStatus::Updating),
       4, uint(UpdateStatus::NewChapters),
       5, uint(UpdateStatus::Error),

       0        // eod
};

void UpdateStatus::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    Q_UNUSED(_o);
    Q_UNUSED(_id);
    Q_UNUSED(_c);
    Q_UNUSED(_a);
}

const QMetaObject UpdateStatus::staticMetaObject = {
    { &QObject::staticMetaObject, qt_meta_stringdata_UpdateStatus.data,
      qt_meta_data_UpdateStatus,  qt_static_metacall, Q_NULLPTR, Q_NULLPTR}
};


const QMetaObject *UpdateStatus::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *UpdateStatus::qt_metacast(const char *_clname)
{
    if (!_clname) return Q_NULLPTR;
    if (!strcmp(_clname, qt_meta_stringdata_UpdateStatus.stringdata0))
        return static_cast<void*>(const_cast< UpdateStatus*>(this));
    return QObject::qt_metacast(_clname);
}

int UpdateStatus::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = QObject::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    return _id;
}
struct qt_meta_stringdata_UpdateInfoModel_t {
    QByteArrayData data[6];
    char stringdata0[44];
};
#define QT_MOC_LITERAL(idx, ofs, len) \
    Q_STATIC_BYTE_ARRAY_DATA_HEADER_INITIALIZER_WITH_OFFSET(len, \
    qptrdiff(offsetof(qt_meta_stringdata_UpdateInfoModel_t, stringdata0) + ofs \
        - idx * sizeof(QByteArrayData)) \
    )
static const qt_meta_stringdata_UpdateInfoModel_t qt_meta_stringdata_UpdateInfoModel = {
    {
QT_MOC_LITERAL(0, 0, 15), // "UpdateInfoModel"
QT_MOC_LITERAL(1, 16, 6), // "qmlGet"
QT_MOC_LITERAL(2, 23, 0), // ""
QT_MOC_LITERAL(3, 24, 3), // "row"
QT_MOC_LITERAL(4, 28, 6), // "column"
QT_MOC_LITERAL(5, 35, 8) // "roleName"

    },
    "UpdateInfoModel\0qmlGet\0\0row\0column\0"
    "roleName"
};
#undef QT_MOC_LITERAL

static const uint qt_meta_data_UpdateInfoModel[] = {

 // content:
       7,       // revision
       0,       // classname
       0,    0, // classinfo
       1,   14, // methods
       0,    0, // properties
       0,    0, // enums/sets
       0,    0, // constructors
       0,       // flags
       0,       // signalCount

 // methods: name, argc, parameters, tag, flags
       1,    3,   19,    2, 0x02 /* Public */,

 // methods: parameters
    QMetaType::QVariant, QMetaType::Int, QMetaType::Int, QMetaType::QString,    3,    4,    5,

       0        // eod
};

void UpdateInfoModel::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    if (_c == QMetaObject::InvokeMetaMethod) {
        UpdateInfoModel *_t = static_cast<UpdateInfoModel *>(_o);
        Q_UNUSED(_t)
        switch (_id) {
        case 0: { QVariant _r = _t->qmlGet((*reinterpret_cast< int(*)>(_a[1])),(*reinterpret_cast< int(*)>(_a[2])),(*reinterpret_cast< const QString(*)>(_a[3])));
            if (_a[0]) *reinterpret_cast< QVariant*>(_a[0]) = _r; }  break;
        default: ;
        }
    }
}

const QMetaObject UpdateInfoModel::staticMetaObject = {
    { &NotifiableModel::staticMetaObject, qt_meta_stringdata_UpdateInfoModel.data,
      qt_meta_data_UpdateInfoModel,  qt_static_metacall, Q_NULLPTR, Q_NULLPTR}
};


const QMetaObject *UpdateInfoModel::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *UpdateInfoModel::qt_metacast(const char *_clname)
{
    if (!_clname) return Q_NULLPTR;
    if (!strcmp(_clname, qt_meta_stringdata_UpdateInfoModel.stringdata0))
        return static_cast<void*>(const_cast< UpdateInfoModel*>(this));
    return NotifiableModel::qt_metacast(_clname);
}

int UpdateInfoModel::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = NotifiableModel::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    if (_c == QMetaObject::InvokeMetaMethod) {
        if (_id < 1)
            qt_static_metacall(this, _c, _id, _a);
        _id -= 1;
    } else if (_c == QMetaObject::RegisterMethodArgumentMetaType) {
        if (_id < 1)
            *reinterpret_cast<int*>(_a[0]) = -1;
        _id -= 1;
    }
    return _id;
}
QT_END_MOC_NAMESPACE
