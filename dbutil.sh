#!/bin/bash

list_tables() {
    container_name=$1
    db_type=$2
    db_name=$3
    username=$4
    password=$5

    case $db_type in
    mysql)
        docker exec -it $container_name mysql -u"$username" -p"$password" $db_name -e "SHOW TABLES;"
        ;;
    postgres)
        docker exec -it -e PGPASSWORD="$password" $container_name psql -U "$username" $db_name -c "\dt"
        ;;
    *)
        echo "Unsupported database type: $db_type" >&2
        return 1
        ;;
    esac
}

list_table_contents() {
    container_name=$1
    db_type=$2
    db_name=$3
    table_name=$4
    username=$5
    password=$6

    case $db_type in
    mysql)
        docker exec -it $container_name mysql -u"$username" -p"$password" $db_name -e "SELECT * FROM $table_name;"
        ;;
    postgres)
        docker exec -it -e PGPASSWORD="$password" $container_name psql -U "$username" $db_name -c "SELECT * FROM $table_name;"
        ;;
    *)
        echo "Unsupported database type: $db_type" >&2
        return 1
        ;;
    esac
}

create_database() {
    container_name=$1
    db_type=$2
    db_name=$3
    username=$4
    password=$5

    case $db_type in
    mysql)
        docker exec -it $container_name mysql -u"$username" -p"$password" -e "CREATE DATABASE IF NOT EXISTS $db_name;"
        if [ $? -ne 0 ]; then
            echo "Failed to create MySQL database" >&2
            return 1
        fi
        echo "MySQL database $db_name created successfully (if it didn't already exist)"
        ;;
    postgres)
        docker exec -it -e PGPASSWORD="$password" $container_name psql -U "$username" -c "CREATE DATABASE $db_name;"
        if [ $? -ne 0 ]; then
            echo "Failed to create Postgres database" >&2
            return 1
        fi
        echo "Postgres database $db_name created successfully (if it didn't already exist)"
        ;;
    *)
        echo "Unsupported database type: $db_type" >&2
        return 1
        ;;
    esac
}

show_databases() {
    container_name=$1
    db_type=$2
    username=$3
    password=$4

    case $db_type in
    mysql)
        docker exec -it $container_name mysql -u"$username" -p"$password" -e "SHOW DATABASES;"
        ;;
    postgres)
        docker exec -it -e PGPASSWORD="$password" $container_name psql -U "$username" -c "\l"
        ;;
    *)
        echo "Unsupported database type: $db_type" >&2
        return 1
        ;;
    esac
}

# Function to handle database deletion
delete_database() {
    container_name=$1
    db_type=$2
    db_name=$3
    username=$4
    password=$5

    case $db_type in
    mysql)
        docker exec -it $container_name mysql -u"$username" -p"$password" -e "DROP DATABASE IF EXISTS $db_name;"
        if [ $? -ne 0 ]; then
            echo "Failed to delete MySQL database" >&2
            return 1
        fi
        echo "MySQL database $db_name deleted successfully."
        ;;
    postgres)
        docker exec -it -e PGPASSWORD="$password" $container_name psql -U "$username" -c "DROP DATABASE $db_name;"
        if [ $? -ne 0 ]; then
            echo "Failed to delete Postgres database" >&2
            return 1
        fi
        echo "Postgres database $db_name deleted successfully."
        ;;
    *)
        echo "Unsupported database type: $db_type" >&2
        return 1
        ;;
    esac
}

usage() {
    echo "Usage: $0 <operation> <options>"
    echo "Available operations:"
    echo "  create-db -c <container_name> -d <database_type> -n <database_name> -u <username> -p <password>"
    echo "            Creates a database in the specified container."
    echo "  delete-db -c <container_name> -d <database_type> -n <database_name> -u <username> -p <password>"
    echo "            Deletes a database from the specified container."
}

# Main control logic
if [[ $# -eq 0 ]]; then
    usage
fi

operation=$1
shift # Shift arguments to access options

case $operation in
"create-db")
    while getopts "c:d:n:u:p:" opt; do
        case $opt in
        c)
            container_name=${OPTARG}
            ;;
        d)
            database_type=${OPTARG}
            ;;
        n)
            database_name=${OPTARG}
            ;;
        u)
            username=${OPTARG}
            ;;
        p)
            password=${OPTARG}
            ;;
        ?)
            echo "Wrong argument, $opt skipping it."
            exit 1
            ;;
        esac
    done
    create_database "$container_name" "$database_type" "$database_name" "$username" "$password"
    ;;
"list-tables")
    while getopts "c:d:n:u:p:" opt; do
        case $opt in
        c) container_name=${OPTARG} ;;
        d) database_type=${OPTARG} ;;
        n) database_name=${OPTARG} ;;
        u) username=${OPTARG} ;;
        p) password=${OPTARG} ;;
        ?)
            echo "Wrong argument, $opt skipping it." >&2
            exit 1
            ;;
        esac
    done
    list_tables "$container_name" "$database_type" "$database_name" "$username" "$password"
    ;;
"list-contents")
    while getopts "c:d:n:t:u:p:" opt; do
        case $opt in
        c) container_name=${OPTARG} ;;
        d) database_type=${OPTARG} ;;
        n) database_name=${OPTARG} ;;
        t) table_name=${OPTARG} ;; # Capture table_name
        u) username=${OPTARG} ;;
        p) password=${OPTARG} ;;
        ?)
            echo "Wrong argument, $opt skipping it." >&2
            exit 1
            ;;
        esac
    done
    list_table_contents "$container_name" "$database_type" "$database_name" "$table_name" "$username" "$password"
    ;;
"delete-db")
    while getopts "c:d:n:u:p:" opt; do
        case $opt in
        c)
            container_name=${OPTARG}
            ;;
        d)
            database_type=${OPTARG}
            ;;
        n)
            database_name=${OPTARG}
            ;;
        u)
            username=${OPTARG}
            ;;
        p)
            password=${OPTARG}
            ;;
        ?)
            echo "Wrong argument, $opt skipping it."
            exit 1
            ;;
        esac
    done
    delete_database "$container_name" "$database_type" "$database_name" "$username" "$password"
    ;;
"list-db")
    while getopts "c:d:n:u:p:" opt; do
        case $opt in
        c)
            container_name=${OPTARG}
            ;;
        d)
            database_type=${OPTARG}
            ;;
        u)
            username=${OPTARG}
            ;;
        p)
            password=${OPTARG}
            ;;
        ?)
            echo "Wrong argument, $opt skipping it."
            exit 1
            ;;
        esac
    done
    show_databases "$container_name" "$database_type" "$username" "$password"
    ;;
*)
    echo "Invalid operation: $operation"
    exit 1
    ;;
esac

#########################################################################################
#                                                                                       #
# $ dbutil list-contents -c postgres -d postgres -n theatre -u misen -t movies -p root  #
# $ dbutil list-tables -c postgres -d postgres -n theatre -u misen -p root              #
#                                                                                       #
#########################################################################################
